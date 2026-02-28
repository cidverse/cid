package gitlabreleasepublish

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitlab/gitlabcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGitLabReleasePublishWithChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(gitlabcommon.GitLabTestData(), nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|gitlab.changelog",
		TargetFile: "/my-project/.tmp/gitlab.changelog",
	}).Return(nil, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/gitlab.changelog").Return(`changes ...`, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{}, nil)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://gitlab.com/api/v4/projects/123456/releases", httpmock.NewStringResponder(200, `{
   "tag_name":"v0.3",
   "description":"Super nice release",
   "name":"New release",
   "created_at":"2019-01-03T02:22:45.118Z",
   "released_at":"2019-01-03T02:22:45.118Z"
}`))

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGitLabReleasePublishAutoChangelog(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(gitlabcommon.GitLabTestData(), nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|gitlab.changelog",
		TargetFile: "/my-project/.tmp/gitlab.changelog",
	}).Return(nil, fmt.Errorf("a error of some kind"))
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{}, nil)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://gitlab.com/api/v4/projects/123456/releases", httpmock.NewStringResponder(200, `{
   "tag_name":"v0.3",
   "description":"Super nice release",
   "name":"New release",
   "created_at":"2019-01-03T02:22:45.118Z",
   "released_at":"2019-01-03T02:22:45.118Z"
}`))

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGitLabReleasePublishSelfHosted(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(gitlabcommon.GitLabSelfHostedTestData(), nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|gitlab.changelog",
		TargetFile: "/my-project/.tmp/gitlab.changelog",
	}).Return(nil, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/gitlab.changelog").Return(`changes ...`, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{}, nil)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://gitlab.local/api/v4/projects/123456/releases", httpmock.NewStringResponder(200, `{
   "tag_name":"v0.3",
   "description":"Super nice release",
   "name":"New release",
   "created_at":"2019-01-03T02:22:45.118Z",
   "released_at":"2019-01-03T02:22:45.118Z"
}`))

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
