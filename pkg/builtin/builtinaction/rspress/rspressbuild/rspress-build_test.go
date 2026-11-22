package rspressbuild

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/rspress/rspresscommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
)

func TestRspressBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(rspresscommon.TestModuleData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "npm install",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "npm run build",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("TARCreateV1", "/my-project/doc_build", "/my-project/.tmp/doc-build.tar").Return(nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-docs",
		File:   "/my-project/.tmp/doc-build.tar",
		Type:   "docs",
		Format: "tar",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
