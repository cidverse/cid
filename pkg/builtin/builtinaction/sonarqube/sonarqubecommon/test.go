package sonarqubecommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func TestModuleData() *actionsdk.ProjectExecutionContextV1Response {
	return &actionsdk.ProjectExecutionContextV1Response{
		ProjectDir: "/my-project",
		Config: &actionsdk.ConfigV1Response{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: []*actionsdk.ProjectModule{
			{
				ProjectDir:        "/my-project",
				ModuleDir:         "/my-project",
				Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/go.mod"}},
				Name:              "github.com/cidverse/my-project",
				Slug:              "github-com-cidverse-my-project",
				BuildSystem:       "gomod",
				BuildSystemSyntax: "default",
				Language:          map[string]string{},
				Submodules:        nil,
			},
		},
		Env: map[string]string{
			"NCI_PROJECT_NAME":        "my-project-name",
			"NCI_PROJECT_DESCRIPTION": "my-project-description",
			"NCI_COMMIT_REF_TYPE":     "branch",
			"NCI_COMMIT_REF_NAME":     "main",
			"SONAR_HOST_URL":          "https://sonarcloud.local",
			"SONAR_ORGANIZATION":      "my-org",
			"SONAR_PROJECTKEY":        "my-project-key",
			"SONAR_DEFAULT_BRANCH":    "main",
			"SONAR_TOKEN":             "my-token",
		},
	}
}
