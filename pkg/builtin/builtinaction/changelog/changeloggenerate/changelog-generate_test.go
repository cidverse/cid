package changeloggenerate

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changelogcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/go-vcs/vcsapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func testProjectData(t *testing.T, refName string, templates []string, commitLimit int) *actionsdk.ProjectExecutionContextV1Response {
	data := changelogcommon.TestProjectData()
	if refName != "" {
		data.Env["NCI_COMMIT_REF_NAME"] = refName
	}

	cfgMap := map[string]interface{}{}
	if templates != nil {
		cfgMap["templates"] = templates
	}
	if commitLimit > 0 {
		cfgMap["commitlimit"] = commitLimit
	}
	if len(cfgMap) > 0 {
		cfgJson, err := json.Marshal(cfgMap)
		if err != nil {
			t.Fatalf("failed to marshal test config: %v", err)
		}
		data.Config.Config = string(cfgJson)
	}

	return data
}

func testCommit() *actionsdk.VCSCommit {
	return &actionsdk.VCSCommit{
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
	}
}

func TestResolvePreviousRelease(t *testing.T) {
	release := func(version, tag string) actionsdk.VCSRelease {
		return actionsdk.VCSRelease{
			Version: version,
			Ref:     vcsapi.VCSRef{Type: "tag", Value: tag},
		}
	}

	tests := []struct {
		name        string
		releases    []actionsdk.VCSRelease
		currentRef  string
		expected    actionsdk.VCSRelease
		expectFound bool
	}{
		{
			name: "stable tag with previous stable",
			releases: []actionsdk.VCSRelease{
				release("1.2.0", "v1.2.0"),
				release("1.1.0", "v1.1.0"),
				release("1.0.0", "v1.0.0"),
			},
			currentRef:  "v1.2.0",
			expected:    release("1.1.0", "v1.1.0"),
			expectFound: true,
		},
		{
			name: "prerelease tag with earlier prerelease",
			releases: []actionsdk.VCSRelease{
				release("1.2.0-rc.2", "v1.2.0-rc.2"),
				release("1.2.0-rc.1", "v1.2.0-rc.1"),
				release("1.1.0", "v1.1.0"),
			},
			currentRef:  "v1.2.0-rc.2",
			expected:    release("1.2.0-rc.1", "v1.2.0-rc.1"),
			expectFound: true,
		},
		{
			name: "prerelease tag after stable only",
			releases: []actionsdk.VCSRelease{
				release("1.2.0-rc.1", "v1.2.0-rc.1"),
				release("1.1.0", "v1.1.0"),
				release("1.0.0", "v1.0.0"),
			},
			currentRef:  "v1.2.0-rc.1",
			expected:    release("1.1.0", "v1.1.0"),
			expectFound: true,
		},
		{
			name: "branch ref with stable releases",
			releases: []actionsdk.VCSRelease{
				release("1.0.0", "v1.0.0"),
				release("1.1.0", "v1.1.0"),
			},
			currentRef:  "main",
			expected:    release("1.1.0", "v1.1.0"),
			expectFound: true,
		},
		{
			name: "branch ref with only prereleases",
			releases: []actionsdk.VCSRelease{
				release("1.2.0-rc.1", "v1.2.0-rc.1"),
			},
			currentRef:  "main",
			expected:    release("1.2.0-rc.1", "v1.2.0-rc.1"),
			expectFound: true,
		},
		{
			name:        "no releases at all",
			releases:    []actionsdk.VCSRelease{},
			currentRef:  "v1.0.0",
			expectFound: false,
		},
		{
			name: "unordered release input",
			releases: []actionsdk.VCSRelease{
				release("1.0.0", "v1.0.0"),
				release("1.2.0", "v1.2.0"),
				release("1.1.0", "v1.1.0"),
			},
			currentRef:  "v1.2.0",
			expected:    release("1.1.0", "v1.1.0"),
			expectFound: true,
		},
		{
			name: "current tag lower than all releases",
			releases: []actionsdk.VCSRelease{
				release("1.0.0", "v1.0.0"),
				release("1.1.0", "v1.1.0"),
			},
			currentRef:  "v1.0.0",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := resolvePreviousRelease(tt.releases, tt.currentRef)
			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestChangelogGenerateWithPreviousRelease(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "", nil, 0), nil)
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
	sdk.On("VCSTagsV1").Return([]actionsdk.VCSTag{
		{RefType: "tag", Value: "v1.2.0"},
		{RefType: "tag", Value: "v1.1.0"},
		{RefType: "tag", Value: "v1.0.0"},
	}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "tag/v1.1.0",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		testCommit(),
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

func TestChangelogGenerateBranchRun(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "main", []string{"discord.changelog"}, 0), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{
		{
			Version: "1.1.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"},
		},
		{
			Version: "1.0.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.0.0"},
		},
	}, nil)
	sdk.On("VCSTagsV1").Return([]actionsdk.VCSTag{
		{RefType: "tag", Value: "v1.1.0"},
		{RefType: "tag", Value: "v1.0.0"},
	}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "tag/v1.1.0",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		testCommit(),
	}, nil)
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(request actionsdk.ArtifactUploadRequest) bool {
		return request.File == "discord.changelog" && strings.Contains(request.Content, "***main***")
	})).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestChangelogGenerateFirstRelease(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "", nil, 0), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		testCommit(),
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

func TestChangelogGenerateMissingPreviousTag(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "", nil, 0), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{
		{
			Version: "1.2.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.2.0"},
		},
		{
			Version: "1.1.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"},
		},
	}, nil)
	sdk.On("VCSTagsV1").Return([]actionsdk.VCSTag{
		{RefType: "tag", Value: "v1.2.0"},
	}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "v1.1.0")
}

func TestChangelogGenerateCommitLimit(t *testing.T) {
	actionsdk.JoinSeparator = "/"
	sdk := actionsdk.NewMockSDKClient(t)

	var logMessages []string
	sdk.On("LogV1", mock.MatchedBy(func(request actionsdk.LogV1Request) bool {
		logMessages = append(logMessages, request.Message)
		return true
	})).Return(nil)

	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "", nil, 250), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{
		{
			Version: "1.1.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"},
		},
	}, nil)
	sdk.On("VCSTagsV1").Return([]actionsdk.VCSTag{
		{RefType: "tag", Value: "v1.1.0"},
	}, nil)

	commits := make([]*actionsdk.VCSCommit, 250)
	for i := range commits {
		commits[i] = testCommit()
	}
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "tag/v1.1.0",
		Limit:    250,
	}).Return(commits, nil)
	sdk.On("ArtifactUploadV1", mock.Anything).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)

	assert.Contains(t, logMessages, "commit limit reached, changelog may be truncated")
}

func TestChangelogGenerateDefaultCommitLimit(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(testProjectData(t, "", nil, 0), nil)
	sdk.On("VCSReleasesV1", actionsdk.VCSReleasesRequest{}).Return([]actionsdk.VCSRelease{
		{
			Version: "1.1.0",
			Ref:     vcsapi.VCSRef{Type: "tag", Value: "v1.1.0"},
		},
	}, nil)
	sdk.On("VCSTagsV1").Return([]actionsdk.VCSTag{
		{RefType: "tag", Value: "v1.1.0"},
	}, nil)
	sdk.On("VCSCommitsV1", actionsdk.VCSCommitsRequest{
		FromHash: "hash/abcdef123456",
		ToHash:   "tag/v1.1.0",
		Limit:    1000,
	}).Return([]*actionsdk.VCSCommit{
		testCommit(),
	}, nil)
	sdk.On("ArtifactUploadV1", mock.Anything).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
