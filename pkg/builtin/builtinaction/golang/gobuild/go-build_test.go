package gobuild

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gocommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoModBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gocommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `go get -v -t ./...`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `go build -buildvcs=false -ldflags "-s -w -X main.version={NCI_COMMIT_REF_RELEASE} -X main.commit={NCI_COMMIT_HASH} -X main.date={TIMESTAMP_RFC3339} -X main.status={NCI_REPOSITORY_STATUS}" -o /my-project/.tmp/linux_amd64 .`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CGO_ENABLED": "false",
			"GOOS":        "linux",
			"GOARCH":      "amd64",
			"GOTOOLCHAIN": "local",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module:        "github-com-cidverse-my-project",
		File:          "/my-project/.tmp/linux_amd64",
		Type:          "binary",
		Format:        "go",
		FormatVersion: "",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
