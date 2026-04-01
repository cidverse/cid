package githubreleasepublish

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/jarcoal/httpmock"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubReleasePublishWithChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactDownloadByteArrayV1", actionsdk.ArtifactDownloadByteArrayRequest{
		ID: "root|changelog|github.changelog",
	}).Return(&actionsdk.ArtifactDownloadByteArrayResult{Bytes: []byte("feat: example")}, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `(artifact_type == "binary" || artifact_type == "signature")`}).Return([]*actionsdk.Artifact{
		/*
			{
				BuildID:    "0",
				JobID:      "0",
				ArtifactID: "my-module|binary|linux_amd64",
				Module:     "my-module",
				Name:       "linux_amd64",
				Type:       "binary",
			},
		*/
	}, nil)
	/*
		sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
			ID:         "my-module|binary|linux_amd64",
			TargetFile: "/my-project/.tmp/linux_amd64",
		}).Return(nil, nil)
	*/

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://api.github.com/repos/cidverse-owner/cidverse-name/releases", httpmock.NewStringResponder(200, `{"id": 1217}`))       // draft release
	httpmock.RegisterResponder("PATCH", "https://api.github.com/repos/cidverse-owner/cidverse-name/releases/1217", httpmock.NewStringResponder(200, `{"id": 1217}`)) // publish release

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGithubReleasePublishAutoChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactDownloadByteArrayV1", actionsdk.ArtifactDownloadByteArrayRequest{
		ID: "root|changelog|github.changelog",
	}).Return(nil, fmt.Errorf("a error of some kind"))
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `(artifact_type == "binary" || artifact_type == "signature")`}).Return([]*actionsdk.Artifact{
		/*
			{
				BuildID:    "0",
				JobID:      "0",
				ArtifactID: "my-module|binary|linux_amd64",
				Module:     "my-module",
				Name:       "linux_amd64",
				Type:       "binary",
			},
		*/
	}, nil)
	/*
		sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
			ID:         "my-module|binary|linux_amd64",
			TargetFile: "/my-project/.tmp/linux_amd64",
		}).Return(nil, nil)
	*/

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://api.github.com/repos/cidverse-owner/cidverse-name/releases", httpmock.NewStringResponder(200, `{"id": 1217}`))       // draft release
	httpmock.RegisterResponder("PATCH", "https://api.github.com/repos/cidverse-owner/cidverse-name/releases/1217", httpmock.NewStringResponder(200, `{"id": 1217}`)) // publish release

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
