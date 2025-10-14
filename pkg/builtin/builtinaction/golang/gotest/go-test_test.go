package gotest

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gocommon"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestGoModTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(gocommon.ModuleTestData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go get -v -t ./...`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go test -vet all -cover -covermode=atomic -coverprofile "/my-project/.tmp/cover.out" -parallel=4 -timeout 10s -count=1 -shuffle=on ./...`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/cover.out",
		Module:        "github-com-cidverse-my-project",
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "out",
	}).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go test -coverprofile "/my-project/.tmp/cover.out" -json -covermode=count -parallel=4 -timeout 10s ./...`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
		CaptureOutput: true,
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0, Stdout: "{}"}, nil)
	sdk.On("FileWrite", "/my-project/.tmp/cover.json", []byte("{}")).Return(nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/cover.json",
		Module:        "github-com-cidverse-my-project",
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "json",
	}).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go tool cover -html "/my-project/.tmp/cover.out" -o "/my-project/.tmp/cover.html"`,
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/cover.html",
		Module:        "github-com-cidverse-my-project",
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "html",
	}).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `go-junit-report -in "/my-project/.tmp/cover.json" -parser gojson -out "/my-project/.tmp/junit.xml"`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:   "/my-project/.tmp/junit.xml",
		Module: "github-com-cidverse-my-project",
		Type:   "report",
		Format: "junit",
	}).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `gocover-cobertura "/my-project/.tmp/cover.out" "/my-project/.tmp/cobertura.xml"`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:   "/my-project/.tmp/cobertura.xml",
		Module: "github-com-cidverse-my-project",
		Type:   "report",
		Format: "cobertura",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
