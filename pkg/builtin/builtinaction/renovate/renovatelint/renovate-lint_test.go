package renovatelint

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestRenovateLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectActionDataV1").Return(ModuleData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "renovate-config-validator --strict .",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func ModuleData() *cidsdk.ProjectActionData {
	return &cidsdk.ProjectActionData{
		ProjectDir: "/my-project",
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
	}
}
