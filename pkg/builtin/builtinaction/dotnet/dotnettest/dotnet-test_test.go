package dotnettest

import (
	_ "embed"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestDotNetTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(dotnetcommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `dotnet restore`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `dotnet test --logger:"junit;LogFilePath=/my-project/.tmp/junit.xml;MethodFormat=Class;FailureBodyFormat=Verbose" --collect "Code Coverage;Format=cobertura"`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	sdk.On("FileList", cidsdk.FileRequest{Directory: "/my-project", Extensions: []string{".xml"}}).Return([]cidsdk.File{cidsdk.NewFile("/my-project/TestResults/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml")}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/TestResults/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml",
		Type:   "report",
		Format: "cobertura",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
