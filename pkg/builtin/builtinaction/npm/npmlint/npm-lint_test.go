package npmlint

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestNodeLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(npmcommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/package.json").Return(`{"scripts": {"lint": ""}}`, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "npm install",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "npm run lint",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("FileList", cidsdk.FileRequest{Directory: "/my-project", Extensions: []string{".sarif"}}).Return([]cidsdk.File{cidsdk.NewFile("/my-project/build/reports/eslint.sarif")}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module:        "my-package",
		File:          "/my-project/build/reports/eslint.sarif",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestNodeLintNoScript(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(npmcommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/package.json").Return(`{}`, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
