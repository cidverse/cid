package changeloggenerate

import (
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changelogcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/go-vcs/vcsapi"

	"github.com/stretchr/testify/assert"
)

func TestChangelogGenerateWithPreviousRelease(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(changelogcommon.TestProjectData(), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{
		{
			Version: "1.2.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.2.0"},
		},
		{
			Version: "1.1.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"},
		},
		{
			Version: "1.0.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"},
		},
	}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "tag/v1.1.0",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		{
			HashShort:   "123456a",
			Hash:        "f7331a7bc3a0531cf8aa4c982d7fefefffcbe8bc",
			Message:     "feat: add cool new feature",
			Description: "",
			Author:      actionsdk.VCSAuthor{Name: "A Person", Email: "email@example.com"},
			Committer:   actionsdk.VCSAuthor{Name: "A Person", Email: "email@example.com"},
			Tags:        nil,
			AuthoredAt:  time.Now(),
			CommittedAt: time.Now(),
			Changes:     nil,
			Context:     nil,
		},
	}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:    "github.changelog",
		Content: "## Features\n- add cool new feature\n\n",
		Type:    "changelog",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestChangelogGenerateFirstRelease(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(changelogcommon.TestProjectData(), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		{
			HashShort:   "123456a",
			Hash:        "f7331a7bc3a0531cf8aa4c982d7fefefffcbe8bc",
			Message:     "feat: add cool new feature",
			Description: "",
			Author:      actionsdk.VCSAuthor{Name: "A Person", Email: "email@example.com"},
			Committer:   actionsdk.VCSAuthor{Name: "A Person", Email: "email@example.com"},
			Tags:        nil,
			AuthoredAt:  time.Now(),
			CommittedAt: time.Now(),
			Changes:     nil,
			Context:     nil,
		},
	}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:    "github.changelog",
		Content: "## Features\n- add cool new feature\n\n",
		Type:    "changelog",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
