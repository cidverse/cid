package plangenerate

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/go-ptr"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

func GeneratePlan(modules []*analyzerapi.ProjectModule, registry catalog.Config, projectDir string, env map[string]string) (Plan, error) {
	planContext := PlanContext{
		ProjectDir:  projectDir,
		Environment: env,
		Registry:    registry,
		Modules:     modules,
	}
	ruleContext := rules.GetRuleContext(env)

	// select workflow
	workflow, err := selectWorkflow(planContext, ruleContext)
	if err != nil {
		return Plan{}, err
	}

	// collect all actions
	actions, err := getWorkflowActions(workflow)
	if err != nil {
		return Plan{}, err
	}

	// generate plan
	steps, err := generateFlatExecutionPlan(planContext, actions)
	if err != nil {
		return Plan{}, err
	}

	// sort steps topologically
	steps, err = SortSteps(steps)
	if err != nil {
		return Plan{}, err
	}

	// collect stages
	var stages []string
	for _, step := range steps {
		if !slices.Contains(stages, step.Stage) {
			stages = append(stages, step.Stage)
		}
	}

	return Plan{
		Name:   workflow.Name,
		Stages: stages,
		Steps:  steps,
	}, nil
}

func generateFlatExecutionPlan(context PlanContext, actions []catalog.WorkflowAction) ([]Step, error) {
	var steps []Step

	// map artifact types to actions that produce them
	artifactProducers := make(map[string][]string)
	for _, action := range actions {
		catalogAction := ptr.Value(context.Registry.FindAction(action.ID))
		for _, artifact := range catalogAction.Metadata.Output.Artifacts {
			if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
				artifactProducers[artifact.Key()] = append(artifactProducers[artifact.Key()], action.ID)
			}
		}
	}

	// create steps for each action, respecting dependencies
	for _, action := range actions {
		catalogAction := ptr.Value(context.Registry.FindAction(action.ID))
		ctx := api.GetActionContext(context.Modules, context.ProjectDir, context.Environment, catalogAction.Metadata.Access)

		var dependencies []string

		// Check required artifacts and add dependencies
		for _, artifact := range catalogAction.Metadata.Input.Artifacts {
			if producers, exists := artifactProducers[artifact.Key()]; exists {
				dependencies = append(dependencies, producers...)
			}
		}

		// Create steps without stage grouping, but store the stage name
		if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
			ruleContext := rules.GetProjectRuleContext(ctx.Env, ctx.Modules)
			if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
				steps = append(steps, Step{
					ID:       strconv.Itoa(len(steps)),
					Name:     catalogAction.Metadata.Name,
					Stage:    action.Stage,
					Scope:    catalogAction.Metadata.Scope,
					Action:   fmt.Sprintf("%s/%s", catalogAction.Repository, catalogAction.Metadata.Name),
					RunAfter: dependencies,
					Order:    1,
					Config:   action.Config,
				})
			}
		} else if catalogAction.Metadata.Scope == catalog.ActionScopeModule {
			for _, m := range ctx.Modules {
				moduleRef := ptr.Value(m)

				ruleContext := rules.GetModuleRuleContext(ctx.Env, &moduleRef)
				if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
					steps = append(steps, Step{
						ID:       strconv.Itoa(len(steps)),
						Name:     fmt.Sprintf("%s - %s", catalogAction.Metadata.Name, moduleRef.Name),
						Stage:    action.Stage,
						Scope:    catalogAction.Metadata.Scope,
						Module:   moduleRef.ID,
						Action:   fmt.Sprintf("%s/%s", catalogAction.Repository, catalogAction.Metadata.Name),
						RunAfter: dependencies,
						Order:    1,
						Config:   action.Config,
					})
				}
			}
		}
	}

	return steps, nil
}
