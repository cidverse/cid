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
		Command: `dotnet test --no-restore --results-directory "/my-project/.tmp/dotnet-test" --logger:"trx" --collect "Code Coverage;Format=cobertura"`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project/.tmp/dotnet-test", Extensions: []string{".trx", ".xml"}}).Return([]actionsdk.File{
		actionsdk.NewFile("/my-project/.tmp/dotnet-test/user_HOST_2025-04-27_18_57_11.trx"),
		actionsdk.NewFile("/my-project/.tmp/dotnet-test/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml"),
	}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/dotnet-test/user_HOST_2025-04-27_18_57_11.trx",
		Type:   "report",
		Format: "trx",
	}).Return("", "", nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/dotnet-test/b6fffe59-5cca-4fe0-9af9-13ec40efcfce/user_HOST_2025-04-27.18_57_11.cobertura.xml",
		Type:   "report",
		Format: "cobertura",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestDotNetTestJunitLogger(t *testing.T) {
	sdk := common.TestSetup(t)
	data := dotnetcommon.ModuleTestData()
	data.Module.Files = []string{"/my-project/tests/tests.csproj"}
	sdk.On("ModuleExecutionContextV1").Return(data, nil)
	sdk.On("FileReadV1", "/my-project/tests/tests.csproj").Return(`<PackageReference Include="JunitXml.TestLogger" Version="6.1.0" />`, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet restore`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `dotnet test --no-restore --results-directory "/my-project/.tmp/dotnet-test" --logger:"trx" --collect "Code Coverage;Format=cobertura" --logger:"junit;LogFilePath=/my-project/.tmp/dotnet-test/{assembly}.junit.xml;MethodFormat=Class;FailureBodyFormat=Verbose"`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project/.tmp/dotnet-test", Extensions: []string{".trx", ".xml"}}).Return([]actionsdk.File{
		actionsdk.NewFile("/my-project/.tmp/dotnet-test/tests.junit.xml"),
	}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/.tmp/dotnet-test/tests.junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
