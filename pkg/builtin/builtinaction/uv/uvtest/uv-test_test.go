package uvtest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestPythonUVPyTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(uvcommon.TestModuleData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `uv run pytest -v --junit-xml=".tmp/pytest.junit.xml"`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   ".tmp/pytest.junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestPythonUVPyTestCoverage(t *testing.T) {
	moduleData := uvcommon.TestModuleData()
	*moduleData.Module.Dependencies = append(*moduleData.Module.Dependencies, cidsdk.ProjectDependency{
		Type: "pypi",
		Id:   "pytest-cov",
	})

	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(moduleData, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `uv run pytest -v --cov --cov-report term --cov-report xml:".tmp/pytest.coverage.xml" --junit-xml=".tmp/pytest.junit.xml"`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   ".tmp/pytest.coverage.xml",
		Type:   "report",
		Format: "cobertura",
	}).Return(nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   ".tmp/pytest.junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
