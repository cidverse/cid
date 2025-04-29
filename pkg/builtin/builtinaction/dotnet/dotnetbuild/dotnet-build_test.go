package dotnetbuild

import (
	_ "embed"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestDotNetBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(dotnetcommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `dotnet restore`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `dotnet build --runtime linux-x64 --self-contained --configuration Release`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
