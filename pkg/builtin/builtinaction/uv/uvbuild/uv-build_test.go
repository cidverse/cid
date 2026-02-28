package uvbuild

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequirementsBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(uvcommon.TestModuleData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "uv build",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
