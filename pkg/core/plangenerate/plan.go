package plangenerate

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/go-ptr"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

func GeneratePlan(modules []*analyzerapi.ProjectModule, registry catalog.Config, projectDir string, env map[string]string, executables []executable.Executable, pinVersions bool) (Plan, error) {
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
	steps, err := generateFlatExecutionPlan(planContext, actions, executables, pinVersions)
	if err != nil {
		return Plan{}, err
	}

	// determine dependencies
	steps = assignStepDependencies(steps, planContext)

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

func generateFlatExecutionPlan(context PlanContext, actions []catalog.WorkflowAction, executables []executable.Executable, pinVersions bool) ([]Step, error) {
	var steps []Step

	// create steps for each action, respecting dependencies
	for _, action := range actions {
		catalogAction := ptr.Value(context.Registry.FindAction(action.ID))
		ctx := api.GetActionContext(context.Modules, context.ProjectDir, context.Environment, catalogAction.Metadata.Access)

		executableConstraints := make(map[string]string)
		for _, ex := range catalogAction.Metadata.Access.Executables {
			versionConstraint := ex.Constraint
			if versionConstraint == "" {
				versionConstraint = executable.AnyVersionConstraint
			}

			// exact version constraints
			if pinVersions {
				c := executable.SelectCandidate(executables, executable.CandidateFilter{
					Types:             nil,
					Executable:        ex.Name,
					VersionPreference: executable.PreferHighest,
					VersionConstraint: versionConstraint,
				})
				if c != nil {
					versionConstraint = fmt.Sprintf("= %s", ptr.Value(c).GetVersion())
				}
			}

			executableConstraints[ex.Name] = versionConstraint
		}

		// create steps without stage grouping, but store the stage name
		if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
			ruleContext := rules.GetProjectRuleContext(ctx.Env, ctx.Modules)
			if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
				steps = append(steps, Step{
					ID:                    strconv.Itoa(len(steps)),
					Name:                  catalogAction.Metadata.Name,
					Stage:                 action.Stage,
					Scope:                 catalogAction.Metadata.Scope,
					Action:                catalogAction.URI,
					RunAfter:              []string{},
					ExecutableConstraints: executableConstraints,
					Order:                 1,
					Config:                action.Config,
				})
			}
		} else if catalogAction.Metadata.Scope == catalog.ActionScopeModule {
			for _, m := range ctx.Modules {
				moduleRef := ptr.Value(m)

				ruleContext := rules.GetModuleRuleContext(ctx.Env, &moduleRef)
				if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
					steps = append(steps, Step{
						ID:                    strconv.Itoa(len(steps)),
						Name:                  fmt.Sprintf("%s - %s", catalogAction.Metadata.Name, moduleRef.Name),
						Stage:                 action.Stage,
						Scope:                 catalogAction.Metadata.Scope,
						Module:                moduleRef.ID,
						Action:                catalogAction.URI,
						RunAfter:              []string{},
						ExecutableConstraints: executableConstraints,
						Order:                 1,
						Config:                action.Config,
					})
				}
			}
		}
	}

	return steps, nil
}

func assignStepDependencies(steps []Step, context PlanContext) []Step {
	actionInstances := make(map[string][]string)   // Track instances of each action
	artifactProducers := make(map[string][]string) // Track artifact producers

	// track action instances and artifact producers
	for _, step := range steps {
		catalogAction := ptr.Value(context.Registry.FindAction(step.Action))

		// track action instances by ID
		actionInstances[step.Action] = append(actionInstances[step.Action], step.Name)

		// track which actions produce which artifacts
		for _, artifact := range catalogAction.Metadata.Output.Artifacts {
			if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
				artifactProducers[artifact.Key()] = append(artifactProducers[artifact.Key()], step.Name)
			}
		}
	}

	// assign dependencies
	for i, step := range steps {
		catalogAction := ptr.Value(context.Registry.FindAction(step.Action))
		var dependencies []string

		// add dependencies based on required artifacts
		for _, artifact := range catalogAction.Metadata.Input.Artifacts {
			if producers, exists := artifactProducers[artifact.Key()]; exists {
				dependencies = append(dependencies, producers...)
			}
		}

		// add dependencies based on explicit `RunAfter`
		for _, requiredAction := range step.RunAfter {
			if instances, exists := actionInstances[requiredAction]; exists {
				dependencies = append(dependencies, instances...)
			}
		}

		steps[i].RunAfter = dependencies
	}

	return steps
}
