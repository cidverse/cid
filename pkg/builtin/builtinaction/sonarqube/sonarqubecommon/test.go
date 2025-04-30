package sonarqubecommon

import cidsdk "github.com/cidverse/cid-sdk-go"

func TestModuleData() *cidsdk.ProjectActionData {
	return &cidsdk.ProjectActionData{
		ProjectDir: "/my-project",
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: []cidsdk.ProjectModule{
			{
				ProjectDir:        "/my-project",
				ModuleDir:         "/my-project",
				Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/go.mod"}},
				Name:              "github.com/cidverse/my-project",
				Slug:              "github-com-cidverse-my-project",
				BuildSystem:       "gomod",
				BuildSystemSyntax: "default",
				Language:          &map[string]string{},
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
