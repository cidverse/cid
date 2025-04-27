package golangbuild

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/golangcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestGoModBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(golangcommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go get -v -t ./...`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go build -buildvcs=false -ldflags "-s -w -X main.version={NCI_COMMIT_REF_RELEASE} -X main.commit={NCI_COMMIT_HASH} -X main.date={TIMESTAMP_RFC3339} -X main.status={NCI_REPOSITORY_STATUS}" -o /my-project/.tmp/linux_amd64 .`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CGO_ENABLED": "false",
			"GOOS":        "linux",
			"GOARCH":      "amd64",
			"GOTOOLCHAIN": "local",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module:        "github-com-cidverse-my-project",
		File:          "/my-project/.tmp/linux_amd64",
		Type:          "binary",
		Format:        "go",
		FormatVersion: "",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
