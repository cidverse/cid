package golangcilint

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gocommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestGoModLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(gocommon.ModuleTestData(), nil)
	sdk.On("FileExists", "/my-project/.golangci.yml").Return(true)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `golangci-lint run --output.text.path stdout --output.sarif.path "/my-project/.tmp/golangci-lint.sarif.json" --issues-exit-code 0`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/golangci-lint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
