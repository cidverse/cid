package helmfiledeploy

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helmfile/helmfilecommon"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/go-playground/validator/v10"
)

const URI = "builtin://actions/helmfile-deploy"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	DeploymentNamespace   string `json:"deployment_namespace"   env:"DEPLOYMENT_NAMESPACE"   validate:"required"`
	DeploymentEnvironment string `json:"deployment_environment" env:"DEPLOYMENT_ENVIRONMENT" validate:"required"`
	HelmfileArgs          string `json:"helmfile_args"          env:"HELMFILE_ARGS"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "helmfile-deploy",
		Description: "Deploy a module using helmfile.",
		Category:    "deploy",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_DEPLOYMENT_TYPE == "helmfile" && CID_WORKFLOW_TYPE == "release"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "DEPLOYMENT_NAMESPACE",
					Description: "The namespace the deployment should be created in",
				},
				{
					Name:        "KUBECONFIG_BASE64",
					Description: "The base 64 encoded Kubernetes config",
				},
				{
					Name:        "KUBECONFIG_.*_BASE64",
					Description: "Environment-specific base 64 encoded Kubernetes config",
					Pattern:     true,
				},
				{
					Name:        "HELMFILE_ARGS",
					Description: "Additional arguments to pass to the helmfile command",
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "helmfile",
					Constraint: helmfilecommon.HelmfileVersionConstraint,
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// defaults
	if cfg.DeploymentNamespace == "" {
		cfg.DeploymentNamespace = d.Env["NCI_PROJECT_UID"] // fallback to project uid
	}
	if cfg.DeploymentEnvironment == "" && d.Deployment.DeploymentEnvironment != "" {
		cfg.DeploymentEnvironment = d.Deployment.DeploymentEnvironment
	}

	// validate
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// prepare kubeconfig
	kubeConfigFile := cidsdk.JoinPath(d.Config.TempDir, "kube", "kubeconfig")
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "Starting Helmfile deployment...", Context: map[string]interface{}{"KUBECONFIG": kubeConfigFile}})
	err = helmcommon.PrepareKubeConfig(kubeConfigFile, d.Deployment.DeploymentEnvironment, d.Env)
	if err != nil {
		return err
	}

	// target cluster
	targetCluster, err := helmcommon.ParseKubeConfigCluster(kubeConfigFile)
	if err != nil {
		return err
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "Target cluster", Context: map[string]interface{}{"name": targetCluster.Name, "api": targetCluster.Cluster.Server, "namespace": cfg.DeploymentNamespace}})

	// init
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `helmfile init --force`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	// deployment
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "Deploying to cluster...", Context: map[string]interface{}{"cluster": targetCluster.Name, "namespace": cfg.DeploymentNamespace, "environment": cfg.DeploymentEnvironment}})
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`helmfile apply -f %q --namespace=%q --environment=%q --suppress-diff %s`, d.Deployment.DeploymentFile, cfg.DeploymentNamespace, d.Deployment.DeploymentEnvironment, cfg.HelmfileArgs),
		WorkDir: d.Module.ModuleDir,
		Env: map[string]string{
			"KUBECONFIG": kubeConfigFile,
		},
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	return nil
}
