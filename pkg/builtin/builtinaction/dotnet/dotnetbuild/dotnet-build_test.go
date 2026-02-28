package dotnetbuild

import (
	_ "embed"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotNetBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(dotnetcommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet restore`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet build --runtime linux-x64 --self-contained --configuration Release`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
