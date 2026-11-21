package builtinworkflow

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/checksum/checksumartifacts"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/sign/signartifacts"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/sign/sigstoresign"
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

func TestReleaseBinaryActionsOnlyRunForBinaryBuildSystems(t *testing.T) {
	for _, actionID := range []string{checksumartifacts.URI, signartifacts.URI, sigstoresign.URI} {
		t.Run(actionID, func(t *testing.T) {
			actionRules := workflowActionRules(t, actionID)
			matches := func(workflowType string, modules []*analyzerapi.ProjectModule) bool {
				ctx := rules.GetProjectRuleContext(map[string]string{}, modules)
				ctx["CID_WORKFLOW_TYPE"] = workflowType
				return rules.AnyRuleMatches(actionRules, ctx)
			}

			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "cargo"}}))
			assert.True(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "maven"}, {BuildSystem: "gomod"}}))
			// nothing to checksum / sign without a binary producing module
			assert.False(t, matches("release", nil))
			assert.False(t, matches("release", []*analyzerapi.ProjectModule{{BuildSystem: "maven"}}))
			// release workflows only
			assert.False(t, matches("main", []*analyzerapi.ProjectModule{{BuildSystem: "gomod"}}))
		})
	}
}
