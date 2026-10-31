package appgithub

import (
	"testing"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/core/plangenerate"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func testWorkflowData() *appconfig.WorkflowData {
	return &appconfig.WorkflowData{
		Version:     "0.11.0",
		Name:        "release",
		JobTimeout:  10,
		WorkflowKey: "release",
		WorkflowConfig: appconfig.WorkflowConfig{
			Type:            "push",
			TriggerManual:   true,
			TriggerPush:     true,
			TriggerPushTags: []string{"v[0-9]+.[0-9]+.[0-9]+"},
		},
		Plan: plangenerate.Plan{
			Name: "release",
			Steps: []plangenerate.Step{
				{Name: "changelog-generate", Slug: "changelog-generate"},
			},
		},
		WorkflowDependency: map[string]dependency.Dependency{
			"pkg:github/cidverse/cid":                             {Id: "github.com/cidverse/cid", Version: "0.11.0", Hash: "abc", GPGFingerprint: "def"},
			"pkg:github-action/actions/checkout":                  {Id: "actions/checkout", Version: "v6.0.2", Hash: "abc"},
			"pkg:github-action/actions/download-artifact":         {Id: "actions/download-artifact", Version: "v8.0.1", Hash: "abc"},
			"pkg:github-action/actions/upload-artifact":           {Id: "actions/upload-artifact", Version: "v7.0.1", Hash: "abc"},
			"pkg:github-action/github/codeql-action/upload-sarif": {Id: "github/codeql-action/upload-sarif", Version: "v4", Hash: "abc"},
			"pkg:github-action/step-security/harden-runner":       {Id: "step-security/harden-runner", Version: "v2.19.3", Hash: "abc"},
			"pkg:github-action-runner-image/ubuntu":               {Id: "ubuntu", Version: "24.04"},
		},
		ReferencedWorkflowDependency: make(map[string]dependency.Dependency),
		IgnoreFiles:                  []string{"README.md"},
	}
}

func TestRenderWorkflowCheckoutFetchesTags(t *testing.T) {
	result, err := renderWorkflow(testWorkflowData(), "wf-main.gohtml", "")
	assert.NoError(t, err)
	assert.Contains(t, result.WorkflowContent, "fetch-depth: 0")
	assert.Contains(t, result.WorkflowContent, "fetch-tags: true")
}

func TestRenderWorkflowJobPermissionsIncludeBaselineScopes(t *testing.T) {
	result, err := renderWorkflow(testWorkflowData(), "wf-main.gohtml", "")
	assert.NoError(t, err)

	var doc struct {
		Jobs map[string]struct {
			Permissions map[string]string `yaml:"permissions"`
		} `yaml:"jobs"`
	}
	assert.NoError(t, yaml.Unmarshal([]byte(result.WorkflowContent), &doc))

	job, ok := doc.Jobs["changelog-generate"]
	assert.True(t, ok, "expected job changelog-generate to be rendered")
	assert.Equal(t, "read", job.Permissions["actions"])
	assert.Equal(t, "read", job.Permissions["contents"])
	assert.Equal(t, "write", job.Permissions["id-token"])
}
