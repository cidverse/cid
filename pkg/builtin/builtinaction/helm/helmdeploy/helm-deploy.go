package helmdeploy

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/util"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/go-playground/validator/v10"
	cp "github.com/otiai10/copy"
)

const URI = "builtin://actions/helm-deploy"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	DeploymentChart          string `json:"deployment_chart"            env:"DEPLOYMENT_CHART"           validate:"required"`
	DeploymentChartVersion   string `json:"deployment_chart_version"    env:"DEPLOYMENT_CHART_VERSION"`
	DeploymentChartLocalPath string `json:"deployment_chart_local_path" env:"DEPLOYMENT_CHART_LOCAL_PATH"`
	DeploymentNamespace      string `json:"deployment_namespace"        env:"DEPLOYMENT_NAMESPACE"       validate:"required"`
	DeploymentID             string `json:"deployment_id"               env:"DEPLOYMENT_ID"              validate:"required"`
	DeploymentEnvironment    string `json:"deployment_environment"      env:"DEPLOYMENT_ENVIRONMENT"     validate:"required"`
	HelmArgs                 string `json:"helm_args"                   env:"HELM_ARGS"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "helm-deploy",
		Description: "The Helm Deploy action is used to deploy a Helm chart to a Kubernetes cluster.",
		Documentation: util.TrimLeftEachLine(`
			# Helm Deploy

			The Helm Deploy action is used to deploy a Helm chart to a Kubernetes cluster.
			...
		`),
		Category: "deploy",
		Scope:    cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_DEPLOYMENT_TYPE == "helm"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "DEPLOYMENT_CHART",
					Description: "The Helm chart to deploy",
				},
				{
					Name:        "DEPLOYMENT_CHART_VERSION",
					Description: "The Helm chart version to deploy (deprecated)",
				},
				{
					Name:        "DEPLOYMENT_CHART_LOCAL_PATH",
					Description: "The path to a helm chart in the local filesystem. (cannot be used together with DEPLOYMENT_CHART/DEPLOYMENT_CHART_VERSION)",
					// Deprecated in favor of DEPLOYMENT_CHART,
				},
				{
					Name:        "DEPLOYMENT_NAMESPACE",
					Description: "The namespace the deployment should be created in",
				},
				{
					Name:        "DEPLOYMENT_ID",
					Description: "The unique identifier of the deployment",
				},
				{
					Name:        "DEPLOYMENT_ENVIRONMENT",
					Description: "The environment the deployment is targeting",
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
					Name:        "HELM_ARGS",
					Description: "Additional arguments to pass to the helm command",
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "helm",
					Constraint: helmcommon.HelmVersionConstraint,
				},
			},
		},
	}
}

func (a Action) GetConfig(env map[string]string) (Config, error) {
	cfg := Config{}
	cidsdk.PopulateFromEnv(&cfg, env)

	// defaults
	if cfg.DeploymentNamespace == "" {
		cfg.DeploymentNamespace = env["NCI_PROJECT_UID"] // fallback to project uid
	}
	if cfg.DeploymentID == "" {
		cfg.DeploymentID = cfg.DeploymentEnvironment // fallback to environment
	}
	if cfg.DeploymentChart == "" && cfg.DeploymentChartLocalPath != "" {
		cfg.DeploymentChart = cfg.DeploymentChartLocalPath
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
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d.Env)
	if err != nil {
		return err
	}

	// prepare kubeconfig
	kubeConfigFile := cidsdk.JoinPath(d.Config.TempDir, "kube", "kubeconfig")
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Starting Helm deployment...", Context: map[string]interface{}{"KUBECONFIG": kubeConfigFile}})
	err = helmcommon.PrepareKubeConfig(kubeConfigFile, d.Deployment.DeploymentEnvironment, d.Env)
	if err != nil {
		return err
	}

	// target cluster
	targetCluster, err := helmcommon.ParseKubeConfigCluster(kubeConfigFile)
	if err != nil {
		return err
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Target cluster", Context: map[string]interface{}{"name": targetCluster.Name, "api": targetCluster.Cluster.Server}})

	// query chart information
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Querying Helm chart information", Context: map[string]interface{}{"chart": cfg.DeploymentChart, "chart-version": cfg.DeploymentChartVersion}})
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command:       fmt.Sprintf(`helm show chart --version %q %q`, cfg.DeploymentChartVersion, cfg.DeploymentChart),
		WorkDir:       d.Module.ModuleDir,
		CaptureOutput: true,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}
	chartMetadata := cmdResult.Stdout
	chart, err := helmcommon.ParseChart([]byte(chartMetadata))
	if err != nil {
		return err
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Found Helm chart", Context: map[string]interface{}{"chart-version": chart.Version, "app-version": chart.AppVersion}})

	// properties
	chartsDir := cidsdk.JoinPath(d.Config.TempDir, "helm-charts")
	chartDir := path.Join(chartsDir, chart.Name)
	_ = os.MkdirAll(chartsDir, 0755)
	chartSource := helmcommon.GetChartSource(cfg.DeploymentChart)

	// local dir branch, copy dir and maybe pull requirements, if missing
	if chartSource == helmcommon.ChartSourceOCI || chartSource == helmcommon.ChartSourceRepository {
		// download chart
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Downloading Helm chart", Context: map[string]interface{}{"chart": cfg.DeploymentChart, "chart-version": cfg.DeploymentChartVersion}})
		cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			Command: fmt.Sprintf(`helm pull --untar --destination %q --version %q %q`, chartsDir, cfg.DeploymentChartVersion, cfg.DeploymentChart),
			WorkDir: d.Module.ModuleDir,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
		}
	} else if chartSource == helmcommon.ChartSourceLocal {
		// copy chart
		chartSourceDir, err := filepath.Abs(cfg.DeploymentChart)
		if err != nil {
			return fmt.Errorf("failed to resolve chart path: %w", err)
		}
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Copying Helm chart", Context: map[string]interface{}{"chart-dir": chartSourceDir}})
		if chartSourceDir == "" {
			return fmt.Errorf("chart not found: %s", cfg.DeploymentChart)
		}

		err = cp.Copy(chartSourceDir, chartDir)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported chart source: %s", cfg.DeploymentChart)
	}

	// deploy
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "installing helm chart onto cluster", Context: map[string]interface{}{"chart": cfg.DeploymentChart, "chart-version": cfg.DeploymentChartVersion, "namespace": cfg.DeploymentNamespace, "release": cfg.DeploymentID}})
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: fmt.Sprintf(`helm upgrade --namespace %q --install --disable-openapi-validation %s %q %q`, cfg.DeploymentNamespace, cfg.HelmArgs, cfg.DeploymentID, chartDir),
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
