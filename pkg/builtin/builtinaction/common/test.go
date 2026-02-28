package common

import (
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/mock"
)

func TestSetup(t *testing.T) *actionsdk.MockSDKClient {
	cidsdk.JoinSeparator = "/"
	sdk := actionsdk.NewMockSDKClient(t)
	sdk.On("LogV1", mock.Anything).Return(nil).Maybe()
	return sdk
}

func TestProjectData() *actionsdk.ProjectExecutionContextV1Response {
	return &actionsdk.ProjectExecutionContextV1Response{
		ProjectDir: "/my-project",
		Config: &actionsdk.ConfigV1Response{
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
