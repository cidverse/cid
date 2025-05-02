package codecovupload

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/codecov/codecovcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestCodeCovUpload(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectActionDataV1").Return(codecovcommon.ProjectTestData(), nil)
	sdk.On("ArtifactList", cidsdk.ArtifactListRequest{Query: `artifact_type == "report" && (format == "junit" || format == "cobertura" || format == "jacoco")`}).Return(&[]cidsdk.ActionArtifact{
		{
			BuildID: "0",
			JobID:   "0",
			ID:      "my-module|report|junit|junit.xml",
			Module:  "my-module",
			Name:    "junit.xml",
			Type:    "report",
			Format:  "junit",
		},
		{
			BuildID: "0",
			JobID:   "0",
			ID:      "my-module|report|cobertura|cobertura.xml",
			Module:  "my-module",
			Name:    "cobertura.xml",
			Type:    "report",
			Format:  "cobertura",
		},
	}, nil)
	sdk.On("ArtifactDownload", cidsdk.ArtifactDownloadRequest{
		ID:         "my-module|report|junit|junit.xml",
		TargetFile: "/my-project/.tmp/junit.xml",
	}).Return(nil)
	sdk.On("ArtifactDownload", cidsdk.ArtifactDownloadRequest{
		ID:         "my-module|report|cobertura|cobertura.xml",
		TargetFile: "/my-project/.tmp/cobertura.xml",
	}).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `codecov --disable-telem upload-process --git-service github -r cidverse/normalizeci --commit-sha abcdef123456 --report-type=test_results --build-url https://localhost:8081 --disable-search --file /my-project/.tmp/junit.xml`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `codecov --disable-telem upload-process --git-service github -r cidverse/normalizeci --commit-sha abcdef123456 --report-type=coverage --build-url https://localhost:8081 --disable-search --file /my-project/.tmp/cobertura.xml`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `codecov --disable-telem create-report-results --git-service github -r cidverse/normalizeci --commit-sha abcdef123456`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `codecov --disable-telem send-notifications --git-service github -r cidverse/normalizeci --commit-sha abcdef123456`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
