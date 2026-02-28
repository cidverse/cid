package codecovupload

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/codecov/codecovcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeCovUpload(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(codecovcommon.ProjectTestData(), nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "report" && (format == "junit" || format == "cobertura" || format == "jacoco")`}).Return([]*actionsdk.Artifact{
		{
			BuildID:    "0",
			JobID:      "0",
			ArtifactID: "my-module|report|junit|junit.xml",
			Module:     "my-module",
			Name:       "junit.xml",
			Type:       "report",
			Format:     "junit",
		},
		{
			BuildID:    "0",
			JobID:      "0",
			ArtifactID: "my-module|report|cobertura|cobertura.xml",
			Module:     "my-module",
			Name:       "cobertura.xml",
			Type:       "report",
			Format:     "cobertura",
		},
	}, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "my-module|report|junit|junit.xml",
		TargetFile: "/my-project/.tmp/junit.xml",
	}).Return(nil, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "my-module|report|cobertura|cobertura.xml",
		TargetFile: "/my-project/.tmp/cobertura.xml",
	}).Return(nil, nil)

	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `codecov --disable-telem upload-process --git-service github -r cidverse/normalizeci --commit-sha abcdef123456 --report-type=test_results --build-url https://localhost:8081 --disable-search --file /my-project/.tmp/junit.xml`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `codecov --disable-telem upload-process --git-service github -r cidverse/normalizeci --commit-sha abcdef123456 --report-type=coverage --build-url https://localhost:8081 --disable-search --file /my-project/.tmp/cobertura.xml`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `codecov --disable-telem create-report-results --git-service github -r cidverse/normalizeci --commit-sha abcdef123456`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `codecov --disable-telem send-notifications --git-service github -r cidverse/normalizeci --commit-sha abcdef123456`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"CODECOV_TOKEN": "codecov-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
