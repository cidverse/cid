package codecovcommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func ProjectTestData() *actionsdk.ProjectExecutionContextV1Response {
	return &actionsdk.ProjectExecutionContextV1Response{
		ProjectDir: "/my-project",
		Config: &actionsdk.ConfigV1Response{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: nil,
		Env: map[string]string{
			"NCI_REPOSITORY_KIND":      "git",
			"NCI_REPOSITORY_REMOTE":    "https://gitlab.com/cidverse/normalizeci.git",
			"NCI_REPOSITORY_HOST_TYPE": "github",
			"NCI_COMMIT_REF_NAME":      "v1.2.0",
			"NCI_COMMIT_HASH":          "abcdef123456",
			"NCI_COMMIT_REF_VCS":       "refs/tags/v1.2.0",
			"NCI_PIPELINE_URL":         "https://localhost:8081",
			"NCI_PROJECT_ID":           "123456",
			"NCI_PROJECT_PATH":         "cidverse/normalizeci",
			"NCI_PROJECT_URL":          "https://gitlab.com/cidverse/normalizeci",
			"CODECOV_TOKEN":            "codecov-token",
		},
	}
}
