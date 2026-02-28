package npmlint

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(npmcommon.TestModuleData(), nil)
	sdk.On("FileReadV1", "/my-project/package.json").Return(`{"scripts": {"lint": ""}}`, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "npm install",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "npm run lint",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project", Extensions: []string{".sarif"}}).Return([]actionsdk.File{actionsdk.NewFile("/my-project/build/reports/eslint.sarif")}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module:        "my-package",
		File:          "/my-project/build/reports/eslint.sarif",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestNodeLintNoScript(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(npmcommon.TestModuleData(), nil)
	sdk.On("FileReadV1", "/my-project/package.json").Return(`{}`, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
