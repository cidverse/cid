package helmpublishregistry

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/helm-publish-registry"

type Action struct {
	Sdk actionsdk.SDKClient
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
		Input: cidsdk.ActionInput{
			Artifacts: []cidsdk.ActionArtifactType{
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
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// find charts
	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: `artifact_type == "helm-chart" && format == "tgz"`})
	if err != nil {
		return err
	}

	// publish
	for _, artifact := range artifacts {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "uploading chart", Context: map[string]interface{}{"chart": artifact.Name}})

		// get chart archive
		chartArchive := cidsdk.JoinPath(d.Config.TempDir, artifact.Name)
		_, err = a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: chartArchive,
		})
		if err != nil {
			return fmt.Errorf("failed to load artifact with id %s", artifact.ArtifactID)
		}

		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "uploading chart to registry", Context: map[string]interface{}{"chart": artifact.Name}})
		_, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			Command: fmt.Sprintf(`helm push %s oci://%s`, chartArchive, cfg.OCIRepository),
			WorkDir: d.Module.ProjectDir,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
