package nodetest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/node/nodecommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestNodeTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(nodecommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/package.json").Return(`{"scripts": {"test": ""}}`, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "npm install",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "npm test",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("FileList", cidsdk.FileRequest{Directory: "/my-project", Extensions: []string{".xml"}}).Return([]cidsdk.File{cidsdk.NewFile("/my-project/build/reports/junit.xml")}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   "/my-project/build/reports/junit.xml",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestNodeTestNoScript(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(nodecommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/package.json").Return(`{}`, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
