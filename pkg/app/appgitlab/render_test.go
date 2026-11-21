package appgitlab

import (
	"strings"
	"testing"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
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
	result, err := renderWorkflow(testWorkflowData(), appconfig.RunnerTags{}, "wf-main.gohtml", "", RenderOptions{})
	require.NoError(t, err)

	assert.Contains(t, result.WorkflowContent, `bash .cid/scripts/install.sh "0.11.0" "abc" "def"`)
	assert.Contains(t, result.WorkflowContent, `--state-file ".cid/state-gitlab.json" --state-wf-name "main" --step "go-build"`)
	assert.NotContains(t, result.WorkflowContent, "CID_PLAN_")
	assert.NotContains(t, result.WorkflowContent, "CID_INSTALL_EOF")
}

func TestRenderWorkflowSelfContained(t *testing.T) {
	data := testWorkflowData()
	result, err := renderWorkflow(data, appconfig.RunnerTags{}, "wf-main.gohtml", "", RenderOptions{SelfContained: true})
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

// renderedJobTags renders data and returns the effective tags of every job (after resolving the .default-job merge key).
func renderedJobTags(t *testing.T, data []appconfig.WorkflowData, tags appconfig.RunnerTags) map[string][]string {
	t.Helper()
	result, err := renderWorkflow(data, tags, "wf-main.gohtml", "", RenderOptions{})
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(result.WorkflowContent), &parsed))

	jobTags := make(map[string][]string)
	for key, v := range parsed {
		job, ok := v.(map[string]any)
		if !ok || job["stage"] == nil {
			continue
		}
		var jt []string
		if raw, ok := job["tags"].([]any); ok {
			jt = []string{}
			for _, tag := range raw {
				jt = append(jt, tag.(string))
			}
		}
		jobTags[key] = jt
	}

	return jobTags
}

func testWorkflowDataWithDeploy() []appconfig.WorkflowData {
	data := testWorkflowData()
	data[0].Plan.Stages = append(data[0].Plan.Stages, "deploy")
	data[0].Plan.Steps = append(data[0].Plan.Steps, plangenerate.Step{Name: "helm-deploy", Slug: "helm-deploy", Stage: "deploy", RunAfter: []string{"go-test"}, RunAfterByName: []string{"go-test"}})
	return data
}

func TestRenderWorkflowRunnerTagsDefaultOnly(t *testing.T) {
	jobTags := renderedJobTags(t, testWorkflowDataWithDeploy(), appconfig.RunnerTags{Default: []string{"linux"}})

	assert.Equal(t, []string{"linux"}, jobTags["main/go-build"])
	assert.Equal(t, []string{"linux"}, jobTags["main/helm-deploy"])
}

func TestRenderWorkflowRunnerTagsPerStage(t *testing.T) {
	jobTags := renderedJobTags(t, testWorkflowDataWithDeploy(), appconfig.RunnerTags{
		Default: []string{"linux", "small"},
		Stages: map[string][]string{
			"build":  {"linux", "large"},
			"deploy": {"deploy-prod"},
			"lint":   {"unused"}, // stage without jobs
		},
	})

	assert.Equal(t, []string{"linux", "large"}, jobTags["main/go-build"])
	assert.Equal(t, []string{"linux", "small"}, jobTags["main/go-test"])
	assert.Equal(t, []string{"deploy-prod"}, jobTags["main/helm-deploy"])
}

func TestRenderWorkflowRunnerTagsStageWithoutTags(t *testing.T) {
	jobTags := renderedJobTags(t, testWorkflowDataWithDeploy(), appconfig.RunnerTags{Default: []string{"linux"}, Stages: map[string][]string{"deploy": {}}})

	assert.Equal(t, []string{"linux"}, jobTags["main/go-build"])
	assert.Equal(t, []string{}, jobTags["main/helm-deploy"], "deploy job must override the anchor's tags with an empty list")
}

func TestRenderWorkflowRunnerTagsStageOnly(t *testing.T) {
	jobTags := renderedJobTags(t, testWorkflowDataWithDeploy(), appconfig.RunnerTags{Stages: map[string][]string{"deploy": {"deploy-prod"}}})

	assert.Nil(t, jobTags["main/go-build"])
	assert.Equal(t, []string{"deploy-prod"}, jobTags["main/helm-deploy"])
}

func TestRenderWorkflowEgressFilterVariable(t *testing.T) {
	variables := func(runtime string, egressFilter string) map[string]any {
		data := testWorkflowData()
		data[0].ContainerRuntime = runtime
		result, err := renderWorkflow(data, appconfig.RunnerTags{}, "wf-main.gohtml", "", RenderOptions{EgressFilter: egressFilter})
		require.NoError(t, err)
		var doc struct {
			Variables map[string]any `yaml:"variables"`
		}
		require.NoError(t, yaml.Unmarshal([]byte(result.WorkflowContent), &doc))
		return doc.Variables
	}

	// disabled by default
	assert.NotContains(t, variables("podman", ""), "CID_EGRESS_FILTER")
	assert.NotContains(t, variables("podman", "off"), "CID_EGRESS_FILTER")

	assert.Equal(t, "audit", variables("podman", "audit")["CID_EGRESS_FILTER"])
	assert.Equal(t, "block", variables("podman", "block")["CID_EGRESS_FILTER"])
	assert.NotContains(t, variables("docker", "audit"), "CID_EGRESS_FILTER", "the filter can't see containers in the dind service")
}

func TestRenderWorkflowIDTokens(t *testing.T) {
	data := testWorkflowData()
	data[0].Plan.Steps[0].Access.IDTokens = []actionsdk.ActionAccessIDToken{{Name: "SIGSTORE_ID_TOKEN", Audience: "sigstore"}}
	result, err := renderWorkflow(data, appconfig.RunnerTags{}, "wf-main.gohtml", "", RenderOptions{})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(result.WorkflowContent), &doc))
	build, ok := doc["main/go-build"].(map[string]any)
	require.True(t, ok)
	test, ok := doc["main/go-test"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, map[string]any{"SIGSTORE_ID_TOKEN": map[string]any{"aud": "sigstore"}}, build["id_tokens"])
	assert.NotContains(t, test, "id_tokens")
}
