package cargotest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargocommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCargoTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileWriteV1", ".tmp/nextest.toml", nextestBytes).Return(nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "cargo nextest run --profile=ci --tool-config-file ci:.tmp/nextest.toml",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:   "/my-project/target/nextest/ci/junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
