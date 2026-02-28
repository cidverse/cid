package dotnettest

import (
	_ "embed"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotNetTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(dotnetcommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet restore`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet test --logger:"junit;LogFilePath=/my-project/.tmp/junit.xml;MethodFormat=Class;FailureBodyFormat=Verbose" --logger:"trx;LogFileName=/my-project/.tmp/vstest.trx" --collect "Code Coverage;Format=cobertura"`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/vstest.trx",
		Type:   "report",
		Format: "trx",
	}).Return("", "", nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return("", "", nil)

	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project", Extensions: []string{".xml"}}).Return([]actionsdk.File{actionsdk.NewFile("/my-project/TestResults/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml")}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/TestResults/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml",
		Type:   "report",
		Format: "cobertura",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
