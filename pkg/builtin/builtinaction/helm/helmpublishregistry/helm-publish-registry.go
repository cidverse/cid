package helmpublishregistry

import (
	"fmt"
	"strings"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/helm-publish-registry"

type Action struct {
	Sdk actionsdk.SDKClient
}

type PublishRegistryConfig struct {
	OCIRepository string `json:"helm_oci_repository" env:"HELM_OCI_REPOSITORY"`
	OCIUsername   string `json:"helm_oci_username"   env:"HELM_OCI_USERNAME"`
	OCIPassword   string `json:"helm_oci_password"   env:"HELM_OCI_PASSWORD"`
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "helm-publish-registry",
		Description: "Publishes the helm chart into a OCI registry.",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeModule,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "helm" && ENV["HELM_OCI_REPOSITORY"] != ""`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{
				{
					Name:        "HELM_OCI_REPOSITORY",
					Description: "The OCI repository for the chart publication, e.g. ghcr.io/my-org/charts (the oci:// prefix is optional).",
					Required:    true,
				},
				{
					Name:        "HELM_OCI_USERNAME",
					Description: "The username for the OCI registry, defaults to GITHUB_ACTOR for ghcr.io.",
					Secret:      true,
				},
				{
					Name:        "HELM_OCI_PASSWORD",
					Description: "The password or token for the OCI registry, defaults to GITHUB_TOKEN for ghcr.io.",
					Secret:      true,
				},
				{
					Name:        "GITHUB_ACTOR",
					Description: "The GitHub actor, used to authenticate with ghcr.io.",
				},
				{
					Name:        "GITHUB_TOKEN",
					Description: "The GitHub token, used to authenticate with ghcr.io.",
					Secret:      true,
				},
			},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name:       "helm",
					Constraint: helmcommon.HelmVersionConstraint,
				},
			},
			Resources: []actionsdk.ActionAccessResource{
				actionsdk.ResourcePackages,
			},
		},
		Input: actionsdk.ActionInput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   "helm-chart",
					Format: "tgz",
				},
			},
		},
	}
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg := PublishRegistryConfig{}
	actionsdk.PopulateFromEnv(&cfg, d.Env)
	repository := strings.TrimSuffix(strings.TrimPrefix(cfg.OCIRepository, "oci://"), "/")
	registryHost := strings.SplitN(repository, "/", 2)[0]
	if registryHost == "ghcr.io" {
		if cfg.OCIUsername == "" {
			cfg.OCIUsername = d.Env["GITHUB_ACTOR"]
		}
		if cfg.OCIPassword == "" {
			cfg.OCIPassword = d.Env["GITHUB_TOKEN"]
		}
	}

	// find charts
	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: helmcommon.ChartArtifactQuery(d.Module.Slug)})
	if err != nil {
		return err
	}

	// use a dedicated registry config, to avoid persisting credentials in the user home
	helmEnv := map[string]string{
		"HELM_REGISTRY_CONFIG": actionsdk.JoinPath(d.Config.TempDir, "helm-registry-config.json"),
	}

	// login
	if cfg.OCIUsername != "" && cfg.OCIPassword != "" {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "logging in to registry", Context: map[string]interface{}{"registry": registryHost}})
		cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			Command: fmt.Sprintf(`helm registry login %q --username %q --password %q`, registryHost, cfg.OCIUsername, cfg.OCIPassword),
			WorkDir: d.Module.ProjectDir,
			Env:     helmEnv,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("helm registry login failed, exit code %d", cmdResult.Code)
		}
	}

	// publish
	for _, artifact := range artifacts {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "uploading chart", Context: map[string]interface{}{"chart": artifact.Name}})

		// get chart archive
		chartArchive := actionsdk.JoinPath(d.Config.TempDir, artifact.Name)
		_, err = a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: chartArchive,
		})
		if err != nil {
			return fmt.Errorf("failed to load artifact with id %s: %w", artifact.ArtifactID, err)
		}

		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "uploading chart to registry", Context: map[string]interface{}{"chart": artifact.Name}})
		cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			Command: fmt.Sprintf(`helm push %q %q`, chartArchive, "oci://"+repository),
			WorkDir: d.Module.ProjectDir,
			Env:     helmEnv,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("helm push of chart %s failed, exit code %d", artifact.Name, cmdResult.Code)
		}
	}

	return nil
}
