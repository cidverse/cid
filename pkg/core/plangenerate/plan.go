package plangenerate

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/cidverse/cid/pkg/app/appcommon"
	actionApi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/go-ptr"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

type GeneratePlanRequest struct {
	Modules         []*analyzerapi.ProjectModule        `json:"modules"`
	Registry        catalog.Config                      `json:"registry"`
	ProjectDir      string                              `json:"project_dir"`
	Env             map[string]string                   `json:"env"`
	Executables     []executable.Executable             `json:"executables"`
	PinVersions     bool                                `json:"pin_versions"`
	Environments    map[string]appcommon.VCSEnvironment `json:"environments"`
	WorkflowVariant string                              `json:"workflow_variant"`
}

func GeneratePlan(request GeneratePlanRequest) (Plan, error) {
	planContext := PlanContext{
		ProjectDir:      request.ProjectDir,
		Registry:        request.Registry,
		Environment:     request.Env,
		Stages:          []string{"build", "test", "lint", "scan", "package", "publish", "deploy"},
		VCSEnvironments: request.Environments,
		Modules:         request.Modules,
	}
	ruleContext := rules.GetRuleContext(request.Env)
	ruleContext["VARIANT"] = request.WorkflowVariant

	// lookup environment info via api - TODO: move to a separate function
	if request.Environments == nil {
		var platform api.Platform

		if request.Env["NCI_REPOSITORY_HOST_SERVER"] == "github.com" && request.Env["NCI_REPOSITORY_HOST_TYPE"] == "github" { // github
			ghToken := request.Env["GH_TOKEN"]
			if ghToken != "" {
				p, err := vcsapp.NewPlatform(vcsapp.PlatformConfig{
					GitHubUsername: "oauth2",
					GitHubToken:    ghToken,
				})
				if err != nil {
					return Plan{}, err
				}

				platform = p
			}
		} else if request.Env["NCI_REPOSITORY_HOST_TYPE"] == "gitlab" { // gitlab
			ciJobToken := request.Env["CI_JOB_TOKEN"]
			if ciJobToken != "" {
				p, err := vcsapp.NewPlatform(vcsapp.PlatformConfig{
					GitLabServer:      request.Env["NCI_REPOSITORY_HOST_SERVER"],
					GitLabAccessToken: ciJobToken,
				})
				if err != nil {
					return Plan{}, err
				}

				platform = p
			}
		}
		if platform != nil {
			repo, err := platform.FindRepository(request.Env["NCI_PROJECT_PATH"])
			if err != nil {
				return Plan{}, err
			}

			envs, err := platform.Environments(repo)
			if err != nil {
				return Plan{}, fmt.Errorf("failed to get environments: %w", err)
			}

			environments := make(map[string]appcommon.VCSEnvironment, len(envs))
			for _, e := range envs {
				// fetch environment variables
				vars, err := platform.EnvironmentVariables(repo, e.Name)
				if err != nil {
					return Plan{}, fmt.Errorf("failed to get environment variables: %w", err)
				}

				environments[e.Name] = appcommon.VCSEnvironment{
					Env:  e,
					Vars: vars,
				}
			}

			request.Environments = environments
		} else {
			slog.Debug("cannot enrich environments for plan generation")
		}
	}

	// select workflow
	workflow, err := selectWorkflow(planContext, ruleContext)
	if err != nil {
		return Plan{}, err
	}
	log.Debug().Str("workflow-name", workflow.Name).Msg("selected workflow")

	// collect all actions
	actions, err := getWorkflowActions(workflow)
	if err != nil {
		return Plan{}, err
	}
	log.Debug().Int("actions", len(actions)).Msg("workflow actions loaded")

	// generate plan
	steps, err := generateFlatExecutionPlan(planContext, actions, request.Executables, request.PinVersions)
	if err != nil {
		return Plan{}, err
	}
	log.Debug().Int("steps", len(steps)).Msg("workflow steps generated")

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
		catalogActionPtr := context.Registry.FindAction(action.ID)
		if catalogActionPtr == nil {
			return nil, fmt.Errorf("action [%s] not found in registry", action.ID)
		}
		catalogAction := ptr.Value(catalogActionPtr)
		ctx := actionApi.GetActionContext(context.Modules, context.ProjectDir, context.Environment, catalogAction.Metadata.Access)

		// pin executable constraints
		var executableConstraints []catalog.ActionAccessExecutable
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

			executableConstraints = append(executableConstraints, catalog.ActionAccessExecutable{
				Name:       ex.Name,
				Constraint: versionConstraint,
			})
		}

		// create steps without stage grouping, but store the stage name
		if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
			ruleContext := rules.GetProjectRuleContext(ctx.Env, ctx.Modules)

			// check if the action rules match, if not check again for each environment
			if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
				steps = append(steps, buildStep(catalogAction, action, len(steps), catalogAction.Metadata.Name, "", "", executableConstraints))
			} else {
				for _, env := range context.VCSEnvironments {
					vcsEnv := util.CloneMap(ctx.Env)
					for _, e := range env.Vars {
						if isReservedVariable(e.Name) {
							continue
						}

						if e.IsSecret {
							vcsEnv[e.Name] = "***"
						} else {
							vcsEnv[e.Name] = e.Value
						}
					}

					envRuleContext := rules.GetProjectRuleContext(vcsEnv, ctx.Modules)
					if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), envRuleContext) {
						steps = append(steps, buildStep(catalogAction, action, len(steps), catalogAction.Metadata.Name, "", env.Env.Name, executableConstraints))
					} else {
						log.Debug().Str("action", action.ID).Str("environment", env.Env.Name).Msg("action skipped by environment filter")
					}
				}
			}
		} else if catalogAction.Metadata.Scope == catalog.ActionScopeModule {
			for _, m := range ctx.Modules {
				moduleRef := ptr.Value(m)
				ruleContext := rules.GetModuleRuleContext(ctx.Env, &moduleRef)

				// check if the action rules match, if not check again for each environment
				if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext) {
					steps = append(steps, buildStep(catalogAction, action, len(steps), catalogAction.Metadata.Name, moduleRef.ID, "", executableConstraints))
				} else {
					for _, env := range context.VCSEnvironments {
						vcsEnv := util.CloneMap(ctx.Env)
						for _, e := range env.Vars {
							if isReservedVariable(e.Name) {
								continue
							}

							if e.IsSecret {
								vcsEnv[e.Name] = "***"
							} else {
								vcsEnv[e.Name] = e.Value
							}
						}

						envRuleContext := rules.GetModuleRuleContext(vcsEnv, &moduleRef)
						if rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), envRuleContext) {
							steps = append(steps, buildStep(catalogAction, action, len(steps), catalogAction.Metadata.Name, moduleRef.ID, env.Env.Name, executableConstraints))
						} else {
							log.Debug().Str("action", action.ID).Str("environment", env.Env.Name).Msg("action skipped by environment filter")
						}
					}
				}
			}
		} else {
			return nil, fmt.Errorf("unsupported action scope [%s]: %s", catalogAction.URI, catalogAction.Metadata.Scope)
		}
	}

	return steps, nil
}

func assignStepDependencies(steps []Step, context PlanContext) []Step {
	actionInstances := make(map[string][]string)   // Track instances of each action
	artifactProducers := make(map[string][]string) // Track artifact producers
	stepsByStage := make(map[string][]string)      // Track steps by stage

	// track action instances and artifact producers
	for _, step := range steps {
		stepsByStage[step.Stage] = append(stepsByStage[step.Stage], step.Slug)

		catalogAction := ptr.Value(context.Registry.FindAction(step.Action))
		actionInstances[step.Action] = append(actionInstances[step.Action], step.Slug)

		for _, artifact := range catalogAction.Metadata.Output.Artifacts {
			if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
				artifactProducers[artifact.Key()] = append(artifactProducers[artifact.Key()], step.Slug)
			}
		}
	}

	// create a mapping of stage names to their indices
	stageIndex := map[string]int{}
	for i, stage := range context.Stages {
		stageIndex[stage] = i
	}

	// assign dependencies
	for i, step := range steps {
		catalogAction := ptr.Value(context.Registry.FindAction(step.Action))
		var dependencies []string
		var usesOutputOf []string

		// add dependencies based on required artifacts
		for _, artifact := range catalogAction.Metadata.Input.Artifacts {
			if producers, exists := artifactProducers[artifact.Key()]; exists {
				dependencies = append(dependencies, producers...)
				usesOutputOf = append(usesOutputOf, producers...)
			}
		}

		// add dependencies based on explicit `RunAfter`
		for _, requiredAction := range step.RunAfter {
			if instances, exists := actionInstances[requiredAction]; exists {
				dependencies = append(dependencies, instances...)
				usesOutputOf = append(usesOutputOf, instances...)
			}
		}

		// Stage-based ordering (only for "late" stages)
		if shouldEnforceStageOrdering(step.Stage) {
			if currentStageIndex, exists := stageIndex[step.Stage]; exists {
				for j := 0; j < currentStageIndex; j++ {
					priorStage := context.Stages[j]
					dependencies = append(dependencies, stepsByStage[priorStage]...)
				}
			}
		}

		steps[i].RunAfter = dependencies
		steps[i].UsesOutputOf = usesOutputOf
	}

	return steps
}

func shouldEnforceStageOrdering(stage string) bool {
	switch stage {
	case "publish", "deploy":
		return true
	default:
		return false
	}
}
