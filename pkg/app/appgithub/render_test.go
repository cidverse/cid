package appgithub

import (
	"testing"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/core/actionsdk"
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

	permissions := renderedJobPermissions(t, result.WorkflowContent, "changelog-generate")
	assert.Equal(t, "read", permissions["actions"])
	assert.Equal(t, "read", permissions["contents"])
	assert.Equal(t, "write", permissions["id-token"])
}

func TestRenderWorkflowJobPermissionsOverrideBaseline(t *testing.T) {
	data := testWorkflowData()
	data.Plan.Steps = []plangenerate.Step{
		{Name: "github-release-publish", Slug: "github-release-publish", Access: actionsdk.ActionAccess{Resources: []actionsdk.ActionAccessResource{actionsdk.ResourceReleases}}},
		{Name: "container-publish", Slug: "container-publish", Access: actionsdk.ActionAccess{Resources: []actionsdk.ActionAccessResource{actionsdk.ResourcePackages}}},
		{Name: "zizmor-scan", Slug: "zizmor-scan", Access: actionsdk.ActionAccess{Resources: []actionsdk.ActionAccessResource{actionsdk.ResourceSecurityEvents}}},
		{Name: "github-pr-comment", Slug: "github-pr-comment"},
	}

	result, err := renderWorkflow(data, "wf-main.gohtml", "")
	assert.NoError(t, err)

	releasePermissions := renderedJobPermissions(t, result.WorkflowContent, "github-release-publish")
	assert.Equal(t, "write", releasePermissions["contents"])
	assert.Equal(t, "write", releasePermissions["id-token"])
	assert.NotContains(t, result.WorkflowContent, "contents: read # required: checkout / repository metadata\n      contents: write")

	packagePermissions := renderedJobPermissions(t, result.WorkflowContent, "container-publish")
	assert.Equal(t, "write", packagePermissions["packages"])

	scanPermissions := renderedJobPermissions(t, result.WorkflowContent, "zizmor-scan")
	assert.Equal(t, "write", scanPermissions["security-events"])

	commentPermissions := renderedJobPermissions(t, result.WorkflowContent, "github-pr-comment")
	assert.Equal(t, "write", commentPermissions["pull-requests"])
}

func TestRenderWorkflowJobPermissionsPullRequestDoesNotEscalate(t *testing.T) {
	data := testWorkflowData()
	data.WorkflowConfig.Type = "pull-request"
	data.Plan.Steps = []plangenerate.Step{
		{
			Name: "github-release-publish",
			Slug: "github-release-publish",
			Access: actionsdk.ActionAccess{Resources: []actionsdk.ActionAccessResource{
				actionsdk.ResourceReleases,
				actionsdk.ResourcePackages,
				actionsdk.ResourceSecurityEvents,
			}},
		},
	}

	result, err := renderWorkflow(data, "wf-main.gohtml", "")
	assert.NoError(t, err)

	permissions := renderedJobPermissions(t, result.WorkflowContent, "github-release-publish")
	assert.Equal(t, "read", permissions["contents"])
	assert.NotContains(t, permissions, "packages")
	assert.NotContains(t, permissions, "security-events")
	assert.NotContains(t, permissions, "id-token", "pull requests must not be able to mint an OIDC token")
}

func renderedJobPermissions(t *testing.T, content string, jobSlug string) map[string]string {
	t.Helper()

	var doc struct {
		Jobs map[string]struct {
			Permissions map[string]string `yaml:"permissions"`
		} `yaml:"jobs"`
	}
	assert.NoError(t, yaml.Unmarshal([]byte(content), &doc))

	job, ok := doc.Jobs[jobSlug]
	assert.True(t, ok, "expected job %s to be rendered", jobSlug)
	return job.Permissions
}
