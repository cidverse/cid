package helmbuild

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/helm-build"

type Action struct {
	Sdk actionsdk.SDKClient
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
		RunIfChanged: []string{
			"**/*",
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
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg := BuildConfig{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// restore the charts/ directory based on the Chart.lock file
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
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
	chartFileContent, err := a.Sdk.FileReadV1(chartFile)
	if err != nil {
		return fmt.Errorf("failed to read chart file: %s", err.Error())
	}
	chart, err := helmcommon.ParseChart([]byte(chartFileContent))
	if err != nil {
		return err
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "building chart", Context: map[string]interface{}{"chart-name": chart.Name, "chart-version": chart.Version}})

	// version
	chartVersion := chart.Version
	if chartVersion == "0.0.0" {
		chartVersion = d.Env["NCI_COMMIT_REF_NAME"]
	}

	// pull dependencies, if any are defined in Chart.yaml
	if len(chart.Dependencies) > 0 {
		cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			Command: `helm dependency build .`,
			WorkDir: d.Module.ModuleDir,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
		}
	}

	// package
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `helm package . --version ` + chartVersion + ` --destination ` + d.Config.TempDir,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	// upload charts
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
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
