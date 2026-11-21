package builtinworkflow

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/checksum/checksumartifacts"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/sign/signartifacts"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/sign/sigstoresign"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/syft/syftsbom"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// workflowActionRules returns the rules of an action in the main workflow
func workflowActionRules(t *testing.T, actionID string) []catalog.WorkflowRule {
	t.Helper()
	for _, wf := range GetWorkflows() {
		if wf.Name != "main" {
			continue
		}
		for _, stage := range wf.Stages {
			for _, action := range stage.Actions {
				if action.ID == actionID {
					return action.Rules
				}
			}
		}
	}
	require.Failf(t, "action not found in main workflow", actionID)
	return nil
}

// workflowActionMatches evaluates the main workflow rules of an action for a workflow type and set of modules
func workflowActionMatches(t *testing.T, actionID string, workflowType string, modules []*analyzerapi.ProjectModule) bool {
	t.Helper()
	ctx := rules.GetProjectRuleContext(map[string]string{}, modules)
	ctx["CID_WORKFLOW_TYPE"] = workflowType
	return rules.AnyRuleMatches(workflowActionRules(t, actionID), ctx)
}

func TestChecksumOnlyRunsForBinaryBuildSystems(t *testing.T) {
	matches := func(workflowType string, modules []*analyzerapi.ProjectModule) bool {
		return workflowActionMatches(t, checksumartifacts.URI, workflowType, modules)
	}

	assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
	assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "cargo"}}))
	assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "maven"}, {BuildSystem: "gomod"}}))
	// nothing to checksum without a binary producing module
	assert.False(t, matches("release", nil))
	assert.False(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "maven"}}))
	// release workflows only
	assert.False(t, matches("main", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
}

func TestSignActionsRunForBinaryAndSBOMBuildSystems(t *testing.T) {
	for _, actionID := range []string{signartifacts.URI, sigstoresign.URI} {
		t.Run(actionID, func(t *testing.T) {
			matches := func(workflowType string, modules []*analyzerapi.ProjectModule) bool {
				return workflowActionMatches(t, actionID, workflowType, modules)
			}

			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "cargo"}}))
			// sboms (syft-sbom / build plugins) are signed too
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "maven"}}))
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "gradle"}}))
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "npm"}}))
			// nothing to sign
			assert.False(t, matches("release", nil))
			assert.False(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "helm"}}))
			// release workflows only
			assert.False(t, matches("main", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
		})
	}
}

// TestSignCoversSyftBuildSystems: every module syft-sbom runs for in a release must also enable signing.
func TestSignCoversSyftBuildSystems(t *testing.T) {
	for _, bs := range syftsbom.BuildSystems {
		for _, actionID := range []string{signartifacts.URI, sigstoresign.URI} {
			ctx := rules.GetProjectRuleContext(map[string]string{}, nil)
			ctx["PROJECT_BUILD_SYSTEMS"] = []string{bs}
			ctx["CID_WORKFLOW_TYPE"] = "release"
			assert.True(t, rules.AnyRuleMatches(workflowActionRules(t, actionID), ctx), "%s: %s", actionID, bs)
		}
	}
}
