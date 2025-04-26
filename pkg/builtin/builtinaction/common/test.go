package common

import (
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid-sdk-go/mocks"
)

func TestSetup(t *testing.T) *mocks.SDKClient {
	cidsdk.JoinSeparator = "/"
	sdk := mocks.NewSDKClient(t)
	return sdk
}

func TestProjectData() *cidsdk.ProjectActionData {
	return &cidsdk.ProjectActionData{
		ProjectDir: "/my-project",
		Config: cidsdk.CurrentConfig{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Modules: nil,
		Env: map[string]string{
			"NCI_REPOSITORY_KIND":        "git",
			"NCI_REPOSITORY_REMOTE":      "https://github.com/cidverse/normalizeci.git",
			"NCI_REPOSITORY_URL":         "https://github.com/cidverse/normalizeci",
			"NCI_REPOSITORY_HOST_SERVER": "github.com",
			"NCI_COMMIT_REF_NAME":        "v1.2.0",
			"NCI_COMMIT_HASH":            "abcdef123456",
			"NCI_COMMIT_REF_VCS":         "refs/tags/v1.2.0",
			"NCI_PROJECT_ID":             "123456",
			"NCI_PROJECT_URL":            "https://github.com/cidverse/normalizeci",
		},
	}
}
