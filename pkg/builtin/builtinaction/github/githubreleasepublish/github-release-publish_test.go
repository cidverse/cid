package githubreleasepublish

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubReleasePublishWithChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|github.changelog",
		TargetFile: "/my-project/.tmp/github.changelog",
	}).Return(nil, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{
		{
			BuildID:    "0",
			JobID:      "0",
			ArtifactID: "my-module|binary|linux_amd64",
			Module:     "my-module",
			Name:       "linux_amd64",
			Type:       "binary",
		},
	}, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "my-module|binary|linux_amd64",
		TargetFile: "/my-project/.tmp/linux_amd64",
	}).Return(nil, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `gh release create "v1.2.0" --verify-tag -F "/my-project/.tmp/github.changelog" '/my-project/.tmp/linux_amd64#my-module/linux_amd64'`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GH_TOKEN": "",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGithubReleasePublishAutoChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|github.changelog",
		TargetFile: "/my-project/.tmp/github.changelog",
	}).Return(nil, fmt.Errorf("a error of some kind"))
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{
		{
			BuildID:    "0",
			JobID:      "0",
			ArtifactID: "my-module|binary|linux_amd64",
			Module:     "my-module",
			Name:       "linux_amd64",
			Type:       "binary",
		},
	}, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "my-module|binary|linux_amd64",
		TargetFile: "/my-project/.tmp/linux_amd64",
	}).Return(nil, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `gh release create "v1.2.0" --verify-tag --generate-notes '/my-project/.tmp/linux_amd64#my-module/linux_amd64'`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"GH_TOKEN": "",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
