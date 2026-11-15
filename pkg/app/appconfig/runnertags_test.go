package appconfig

import (
	"testing"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/stretchr/testify/assert"
)

func TestParseRunnerTags(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, ParseRunnerTags(" a, ,b ,"))
	assert.Equal(t, []string{}, ParseRunnerTags(""))
	assert.NotNil(t, ParseRunnerTags(""))
}

func TestStageKey(t *testing.T) {
	assert.Equal(t, "build", StageKey("BUILD"))
	assert.Equal(t, "build", StageKey("build"))
	assert.Equal(t, "my_stage", StageKey("my-stage"))
}

func TestRunnerTagsFromEnviron(t *testing.T) {
	r := RunnerTagsFromEnviron([]string{
		"PATH=/usr/bin",
		"CID_RUNNER_TAGS=linux, small",
		"CID_RUNNER_TAGS_DEPLOY=deploy",
		"CID_RUNNER_TAGS_BUILD=",
		"CID_RUNNER_TAGS_=ignored",
		"CID_RUNNER_TAGSX=ignored",
	})

	assert.Equal(t, RunnerTags{
		Default: []string{"linux", "small"},
		Stages:  map[string][]string{"deploy": {"deploy"}, "build": {}},
	}, r)
}

func TestRunnerTagsForStage(t *testing.T) {
	r := RunnerTags{Default: []string{"linux"}, Stages: map[string][]string{"deploy": {"deploy"}, "build": {}}}

	assert.Equal(t, []string{"linux"}, r.ForStage("test"))
	assert.Equal(t, []string{"deploy"}, r.ForStage("deploy"))
	assert.Equal(t, []string{"deploy"}, r.ForStage("DEPLOY"))
	assert.Equal(t, []string{}, r.ForStage("build"))
	assert.Nil(t, RunnerTags{}.ForStage("build"))
}

func TestRunnerTagsMerge(t *testing.T) {
	base := RunnerTags{Default: []string{"saas"}}
	server := RunnerTagsFromEnviron([]string{"CID_RUNNER_TAGS=linux", "CID_RUNNER_TAGS_DEPLOY=deploy"})

	tests := []struct {
		name     string
		vars     []api.CIVariable
		stage    string
		expected []string
	}{
		{name: "server default", stage: "build", expected: []string{"linux"}},
		{name: "server stage", stage: "deploy", expected: []string{"deploy"}},
		{
			name:     "project default keeps server stage override",
			vars:     []api.CIVariable{{Name: "CID_RUNNER_TAGS", Value: "gpu"}},
			stage:    "deploy",
			expected: []string{"deploy"},
		},
		{
			name:     "project default applies to stages without override",
			vars:     []api.CIVariable{{Name: "CID_RUNNER_TAGS", Value: "gpu"}},
			stage:    "build",
			expected: []string{"gpu"},
		},
		{
			name:     "project stage override",
			vars:     []api.CIVariable{{Name: "CID_RUNNER_TAGS_BUILD", Value: "large"}},
			stage:    "build",
			expected: []string{"large"},
		},
		{
			name:     "project disables tags of a stage",
			vars:     []api.CIVariable{{Name: "CID_RUNNER_TAGS_DEPLOY", Value: ""}},
			stage:    "deploy",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := base.Merge(server).Merge(RunnerTagsFromVariables(tt.vars))
			assert.Equal(t, tt.expected, resolved.ForStage(tt.stage))
		})
	}

	// merging must not modify the inputs
	_ = server.Merge(RunnerTags{Stages: map[string][]string{"build": {"x"}}})
	assert.NotContains(t, server.Stages, "build")
}
