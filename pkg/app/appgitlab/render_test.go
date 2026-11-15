package appgitlab

import (
	"strings"
	"testing"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func testWorkflowData() []appconfig.WorkflowData {
	return []appconfig.WorkflowData{
		{
			Version:          "0.11.0",
			Name:             "Main",
			NameSlug:         "main",
			WorkflowKey:      "main",
			JobTimeout:       10,
			ContainerRuntime: "podman",
			WorkflowConfig: appconfig.WorkflowConfig{
				Type:                "push",
				TriggerPush:         true,
				TriggerPushBranches: []string{"main"},
			},
			Plan: plangenerate.Plan{
				Name:   "main",
				Stages: []string{"build", "test"},
				Steps: []plangenerate.Step{
					{Name: "go-build", Slug: "go-build", Stage: "build"},
					{Name: "go-test", Slug: "go-test", Stage: "test", RunAfter: []string{"go-build"}, RunAfterByName: []string{"go-build"}},
				},
			},
			WorkflowDependency: map[string]dependency.Dependency{
				"pkg:github/cidverse/cid":                         {Id: "github.com/cidverse/cid", Version: "0.11.0", Hash: "abc", GPGFingerprint: "def"},
				"pkg:docker/podman/stable?repository_url=quay.io": {Id: "quay.io/podman/stable", Version: "v5", Hash: "sha256:abc"},
			},
			ReferencedWorkflowDependency: make(map[string]dependency.Dependency),
		},
	}
}

func TestRenderWorkflowDefaultUsesRepositoryFiles(t *testing.T) {
	result, err := renderWorkflow(testWorkflowData(), nil, "wf-main.gohtml", "", RenderOptions{})
	require.NoError(t, err)

	assert.Contains(t, result.WorkflowContent, `bash .cid/scripts/install.sh "0.11.0" "abc" "def"`)
	assert.Contains(t, result.WorkflowContent, `--state-file ".cid/state-gitlab.json" --state-wf-name "main" --step "go-build"`)
	assert.NotContains(t, result.WorkflowContent, "CID_PLAN_")
	assert.NotContains(t, result.WorkflowContent, "CID_INSTALL_EOF")
}

func TestRenderWorkflowSelfContained(t *testing.T) {
	data := testWorkflowData()
	result, err := renderWorkflow(data, nil, "wf-main.gohtml", "", RenderOptions{SelfContained: true})
	require.NoError(t, err)
	content := result.WorkflowContent

	// no dependency on repository files
	assert.NotContains(t, content, ".cid/")
	assert.Contains(t, content, `--plan-env "CID_PLAN_MAIN" --step "go-build"`)

	// valid YAML, plan variable decodes to the workflow plan, install script is inlined as heredoc
	var parsed struct {
		Default struct {
			BeforeScript []string `yaml:"before_script"`
		} `yaml:"default"`
		Variables map[string]any `yaml:"variables"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(content), &parsed))

	planVar, ok := parsed.Variables["CID_PLAN_MAIN"].(map[string]any)
	require.True(t, ok, "CID_PLAN_MAIN must be a variable with value/expand")
	assert.Equal(t, false, planVar["expand"])
	plan, err := plangenerate.DecodePlan(planVar["value"].(string))
	require.NoError(t, err)
	assert.Equal(t, data[0].Plan.Steps[1].Slug, plan.Steps[1].Slug)

	require.Len(t, parsed.Default.BeforeScript, 1)
	install := parsed.Default.BeforeScript[0]
	assert.True(t, strings.HasPrefix(install, `bash -s -- "0.11.0" "abc" "def" <<'CID_INSTALL_EOF'`+"\n#!/usr/bin/env bash\n"), install)
	assert.True(t, strings.HasSuffix(install, "\nCID_INSTALL_EOF\n"), install)
	assert.Contains(t, install, `CID_SHA256="${2:?Error: CID_SHA256 must be provided for integrity check}"`)
}

func TestPlanVariableName(t *testing.T) {
	assert.Equal(t, "CID_PLAN_MAIN", planVariableName("main"))
	assert.Equal(t, "CID_PLAN_PULL_REQUEST", planVariableName("pull-request"))
}
