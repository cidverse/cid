package helmcommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const HelmVersionConstraint = ">=3.16.0"

func GetHelmTestData(env map[string]string, debug bool) *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: "/my-project",
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project/charts/mychart",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/charts/mychart/Chart.yaml"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       string(cidsdk.BuildSystemHelm),
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       debug,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
		Env: env,
		Deployment: &actionsdk.DeploymentV1Response{
			DeploymentSpec:        "deployment-dotenv",
			DeploymentType:        "helm",
			DeploymentEnvironment: "dev",
			DeploymentFile:        "",
			Properties:            map[string]string{},
		},
	}
}

func GetHelmfileTestData(env map[string]string, debug bool) *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: "/my-project",
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/Helmfile.yaml"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       string(cidsdk.BuildSystemHelmfile),
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       debug,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
		Env: env,
		Deployment: &actionsdk.DeploymentV1Response{
			DeploymentSpec:        "deployment-helmfile",
			DeploymentType:        "helmfile",
			DeploymentEnvironment: "dev",
			DeploymentFile:        "helmfile.yaml",
			Properties:            map[string]string{},
		},
	}
}
