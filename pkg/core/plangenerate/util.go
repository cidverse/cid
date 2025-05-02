package plangenerate

import (
	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
)

// selectWorkflow returns the best matching workflow based on the rules
func selectWorkflow(context PlanContext, ruleContext map[string]interface{}) (catalog.Workflow, error) {
	for _, workflow := range context.Registry.Workflows {
		if !rules.AnyRuleMatches(workflow.Rules, ruleContext) {
			continue
		}

		return workflow, nil
	}

	return catalog.Workflow{}, ErrNoSuitableWorkflowFound
}

func getWorkflowActions(workflow catalog.Workflow, ruleContext map[string]interface{}) ([]catalog.WorkflowAction, error) {
	var actions []catalog.WorkflowAction

	for _, stage := range workflow.Stages {
		if !rules.AnyRuleMatches(stage.Rules, ruleContext) {
			continue
		}

		for _, action := range stage.Actions {
			action.Stage = stage.Name
			actions = append(actions, action)
		}
	}

	return actions, nil
}

func isReservedVariable(name string) bool {
	if strings.HasPrefix(name, "NCI_") {
		return true
	}

	return slices.Contains(rules.ReservedVariables, name)
}

func projectEnv(env map[string]string, vars []api.CIVariable) map[string]string {
	result := util.CloneMap(env)

	for _, v := range vars {
		if isReservedVariable(v.Name) {
			continue
		}

		if v.IsSecret {
			result[v.Name] = "***"
		} else {
			result[v.Name] = v.Value
		}
	}

	return result
}

func projectEnvironmentEnv(env map[string]string, vars []api.CIVariable, environment appcommon.VCSEnvironment) map[string]string {
	result := util.CloneMap(env)

	for _, v := range vars {
		if isReservedVariable(v.Name) {
			continue
		}

		if v.IsSecret {
			result[v.Name] = "***"
		} else {
			result[v.Name] = v.Value
		}
	}

	for _, v := range environment.Vars {
		if isReservedVariable(v.Name) {
			continue
		}

		if v.IsSecret {
			result[v.Name] = "***"
		} else {
			result[v.Name] = v.Value
		}
	}

	return result
}
