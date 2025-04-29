package helmpublishregistry

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/helm-publish-registry"

type Action struct {
	Sdk cidsdk.SDKClient
}

type PublishRegistryConfig struct {
	OCIRepository string `json:"helm_oci_repository" env:"HELM_OCI_REPOSITORY"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "helm-publish-registry",
		Description: "Publishes the helm chart into a OCI registry.",
		Category:    "publish",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "helm" && ENV["HELM_OCI_REPOSITORY"] != ""`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "HELM_OCI_REPOSITORY",
					Description: "The url of the OCI registry for the chart publication.",
					Required:    true,
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

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	cfg := PublishRegistryConfig{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// find charts
	artifacts, err := a.Sdk.ArtifactList(cidsdk.ArtifactListRequest{Query: `artifact_type == "helm-chart" && format == "tgz"`})
	if err != nil {
		return err
	}

	// publish
	for _, artifact := range *artifacts {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "uploading chart", Context: map[string]interface{}{"chart": artifact.Name}})

		// get chart archive
		chartArchive := cidsdk.JoinPath(d.Config.TempDir, artifact.Name)
		err = a.Sdk.ArtifactDownload(cidsdk.ArtifactDownloadRequest{
			ID:         artifact.ID,
			TargetFile: chartArchive,
		})
		if err != nil {
			return fmt.Errorf("failed to load artifact with id %s", artifact.ID)
		}

		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "uploading chart to registry", Context: map[string]interface{}{"chart": artifact.Name}})
		_, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: fmt.Sprintf(`helm push %s oci://%s`, chartArchive, cfg.OCIRepository),
			WorkDir: d.Module.ProjectDir,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
