package gitlabcommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

func GitLabTestData() *cidsdk.ProjectActionData {
	return &cidsdk.ProjectActionData{
		ProjectDir: "/my-project",
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: nil,
		Env: map[string]string{
			"NCI_REPOSITORY_KIND":   "git",
			"NCI_REPOSITORY_REMOTE": "https://gitlab.com/cidverse/normalizeci.git",
			"NCI_COMMIT_REF_NAME":   "v1.2.0",
			"NCI_COMMIT_HASH":       "abcdef123456",
			"NCI_COMMIT_REF_VCS":    "refs/tags/v1.2.0",
			"NCI_PROJECT_ID":        "123456",
			"NCI_PROJECT_URL":       "https://gitlab.com/cidverse/normalizeci",
			"CI_JOB_TOKEN":          "dummy-token",
		},
	}
}

func GitLabSelfHostedTestData() *cidsdk.ProjectActionData {
	return &cidsdk.ProjectActionData{
		ProjectDir: "/my-project",
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: nil,
		Env: map[string]string{
			"NCI_REPOSITORY_KIND":   "git",
			"NCI_REPOSITORY_REMOTE": "https://gitlab.local/cidverse/normalizeci.git",
			"NCI_COMMIT_REF_NAME":   "v1.2.0",
			"NCI_COMMIT_HASH":       "abcdef123456",
			"NCI_COMMIT_REF_VCS":    "refs/tags/v1.2.0",
			"NCI_PROJECT_ID":        "123456",
			"NCI_PROJECT_URL":       "https://gitlab.local/cidverse/normalizeci",
			"CI_JOB_TOKEN":          "dummy-token",
		},
	}
}
