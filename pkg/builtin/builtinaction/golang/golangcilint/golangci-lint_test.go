package golangcilint

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gocommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoModLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gocommon.ModuleTestData(), nil)
	sdk.On("FileExistsV1", "/my-project/.golangci.yml").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `golangci-lint run --output.text.path stdout --output.sarif.path "/my-project/.tmp/golangci-lint.sarif.json" --issues-exit-code 0`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/golangci-lint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
