package helmbuild

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/helm-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type BuildConfig struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "helm-build",
		Description: "Builds the helm chart using helm cli.",
		Category:    "build",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "helm"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "helm",
					Constraint: helmcommon.HelmVersionConstraint,
				},
			},
		},
		Output: cidsdk.ActionOutput{
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
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	cfg := BuildConfig{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// restore the charts/ directory based on the Chart.lock file
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `helm dependency build .`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	// parse chart
	chartFile := cidsdk.JoinPath(d.Module.ModuleDir, "Chart.yaml")
	chartFileContent, err := a.Sdk.FileRead(chartFile)
	if err != nil {
		return fmt.Errorf("failed to read chart file: %s", err.Error())
	}
	chart, err := helmcommon.ParseChart([]byte(chartFileContent))
	if err != nil {
		return err
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "building chart", Context: map[string]interface{}{"chart-name": chart.Name, "chart-version": chart.Version}})

	// version
	chartVersion := chart.Version
	if chartVersion == "0.0.0" {
		chartVersion = d.Env["NCI_COMMIT_REF_NAME"]
	}

	// package
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `helm package . --version ` + chartVersion + ` --destination ` + d.Config.TempDir,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	// upload charts
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:   cidsdk.JoinPath(d.Config.TempDir, fmt.Sprintf("%s-%s.tgz", chart.Name, chartVersion)),
		Module: d.Module.Slug,
		Type:   "helm-chart",
		Format: "tgz",
	})
	if err != nil {
		return err
	}

	return nil
}
