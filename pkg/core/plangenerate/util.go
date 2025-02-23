package plangenerate

import (
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

func getWorkflowActions(workflow catalog.Workflow) ([]catalog.WorkflowAction, error) {
	var actions []catalog.WorkflowAction

	for _, stage := range workflow.Stages {
		for _, action := range stage.Actions {
			action.Stage = stage.Name
			actions = append(actions, action)
		}
	}

	return actions, nil
}
