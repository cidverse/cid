package cargotest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargocommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestCargoTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileWrite", ".tmp/nextest.toml", nextestBytes).Return(nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "cargo nextest run --profile=ci --tool-config-file ci:.tmp/nextest.toml",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:   "/my-project/target/nextest/ci/junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
