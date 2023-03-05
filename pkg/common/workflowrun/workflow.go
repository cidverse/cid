package workflowrun

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cidverse/cid/pkg/core/actionexecutor"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/repoanalyzer"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
)

// IsWorkflowExecutable returns true if the workflow is executable (enabled + at least one rule matches)
func IsWorkflowExecutable(w *catalog.Workflow, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(w.Rules, rules.GetRuleContext(env))

	if len(w.Rules) == 0 || matchingRules > 0 {
		return true
	}

	return false
}

// IsStageExecutable returns true if the stage is executable (enabled + at least one rule matches)
func IsStageExecutable(s *catalog.WorkflowStage, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(s.Rules, rules.GetRuleContext(env))

	if len(s.Rules) == 0 || matchingRules > 0 {
		return true
	}

	return false
}

// IsActionExecutable returns true if the action is executable (enabled + at least one rule matches)
func IsActionExecutable(a *catalog.Action, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(a.Rules, rules.GetRuleContext(env))

	if len(a.Rules) == 0 || matchingRules > 0 {
		return true
	}

	return false
}

// FirstWorkflowMatchingRules returns the first workflow that matches at least one rule
func FirstWorkflowMatchingRules(workflows []catalog.Workflow, env map[string]string) *catalog.Workflow {
	// select workflow
	log.Info().Msg("evaluating all workflows")
	for _, wf := range workflows {
		log.Debug().Str("workflow", wf.Name).Msg("evaluating workflow rules")

		if len(wf.Rules) > 0 {
			for _, rule := range wf.Rules {
				match := rules.EvaluateRule(rule, rules.GetRuleContext(env))
				log.Debug().Str("workflow", wf.Name).Bool("match", match).Msg("evaluated rule")
				return &wf
			}
		} else {
			log.Debug().Str("workflow", wf.Name).Msg("workflow match found, no rules")
			return &wf
		}
	}

	return nil
}

func RunWorkflow(cfg *config.CIDConfig, wf *catalog.Workflow, env map[string]string, projectDir string, stagesFilter []string, modulesFilter []string) {
	log.Debug().Str("workflow", wf.Name).Msg("workflow start")
	start := time.Now()
	ruleContext := rules.GetRuleContext(env)

	if rules.AnyRuleMatches(wf.Rules, ruleContext) {
		for i := range wf.Stages {
			if len(stagesFilter) == 0 || funk.Contains(stagesFilter, wf.Stages[i].Name) {
				RunWorkflowStage(cfg, &wf.Stages[i], env, projectDir, modulesFilter)
			} else {
				log.Debug().Str("workflow", wf.Name).Str("stage", wf.Stages[i].Name).Strs("filter", stagesFilter).Msg("stage has been skipped")
			}
		}

		// complete
		log.Info().Str("workflow", wf.Name).Str("duration", time.Since(start).String()).Msg("workflow completed")
	} else {
		log.Debug().Str("workflow", wf.Name).Msg("no workflow rule matches, not running workflow")
	}
}

func RunWorkflowStage(cfg *config.CIDConfig, stage *catalog.WorkflowStage, env map[string]string, projectDir string, modulesFilter []string) {
	log.Debug().Str("stage", stage.Name).Msg("stage start")
	start := time.Now()
	ruleContext := rules.GetRuleContext(env)

	if rules.AnyRuleMatches(stage.Rules, ruleContext) {
		for i := range stage.Actions {
			RunWorkflowAction(cfg, &stage.Actions[i], env, projectDir, modulesFilter)
		}

		// complete
		log.Info().Str("stage", stage.Name).Str("duration", time.Since(start).String()).Msg("stage completed")
	} else {
		log.Debug().Str("stage", stage.Name).Msg("no workflow rule matches, not running stage")
	}
}

func RunWorkflowAction(cfg *config.CIDConfig, action *catalog.WorkflowAction, env map[string]string, projectDir string, modulesFilter []string) {
	log.Debug().Str("action", action.ID).Msg("action start")
	catalogAction := cfg.Registry.FindAction(action.ID)
	if catalogAction == nil {
		log.Fatal().Str("action_id", action.ID).Msg("workflow configuration error, referencing actions that do not exist")
	}
	modules := repoanalyzer.AnalyzeProject(projectDir, filesystem.GetWorkingDirectory())
	ctx := api.GetActionContext(modules, projectDir, env, &catalogAction.Access)

	// serialize action config for pass-thru
	configAsJSON, _ := json.Marshal(&action.Config)
	ctx.Config = string(configAsJSON)

	// project-scoped actions
	if catalogAction.Scope == catalog.ActionScopeProject {
		ruleContext := rules.GetRuleContext(ctx.Env)
		ruleMatch := rules.AnyRuleMatches(append(action.Rules, catalogAction.Rules...), ruleContext)
		log.Debug().Str("Trace", action.ID).Bool("rules_match", ruleMatch).Msg("check action rules")
		if ruleMatch {
			runWorkflowAction(catalogAction, action, &ctx)
		}
	}

	// module-scoped actions
	if catalogAction.Scope == catalog.ActionScopeModule {
		// for each module
		for _, m := range ctx.Modules {
			moduleRef := *m
			log.Trace().Str("action", action.ID).Str("module", moduleRef.Slug).Msg("action for module")

			// customize context
			ctx.CurrentModule = &moduleRef

			// check module filter
			if len(modulesFilter) > 0 && !funk.Contains(modulesFilter, moduleRef.Name) {
				log.Trace().Str("action", action.ID).Str("module", moduleRef.Slug).Strs("filter", modulesFilter).Msg("action skipped by module filter")
				continue
			}

			var ruleContext = rules.GetModuleRuleContext(ctx.Env, &moduleRef)
			ruleMatch := rules.AnyRuleMatches(append(action.Rules, catalogAction.Rules...), ruleContext)
			log.Trace().Str("action", action.ID).Str("module", moduleRef.Name).Bool("rules_match", ruleMatch).Msg("check action rules")
			if ruleMatch {
				runWorkflowAction(catalogAction, action, &ctx)
			}
		}
	}

	log.Debug().Str("action", action.ID).Msg("action end")
}

func runWorkflowAction(catalogAction *catalog.Action, action *catalog.WorkflowAction, ctx *api.ActionExecutionContext) {
	start := time.Now()
	ruleContext := rules.GetRuleContext(ctx.Env)
	if rules.AnyRuleMatches(action.Rules, ruleContext) {
		log.Info().Str("action", action.ID).Msg("action start")
		stateFile := filepath.Join(ctx.Paths.Artifact, "state.json")

		// state: retrieve/init
		localState := state.GetStateFromFile(stateFile)
		localState.Modules = ctx.Modules

		// add action to log
		localState.AuditLog = append(localState.AuditLog, state.AuditEvents{
			Timestamp: time.Now(),
			Type:      "action",
			Payload: map[string]string{
				"action": catalogAction.Repository + "/" + catalogAction.Name + "@" + catalogAction.Version,
				"uri":    fmt.Sprintf("oci://%s", catalogAction.Container.Image),
			},
		})

		// serialize action config for pass-thru
		actConfig, _ := yaml.Marshal(&action.Config)
		log.Trace().Str("action", action.ID).Str("type", string(catalogAction.Type)).Str("config", string(actConfig)).Msg("action configuration")

		// paths
		filesystem.CreateDirectory(ctx.Paths.Temp)
		filesystem.CreateDirectory(ctx.Paths.Artifact)

		// execute
		actionExecutor := actionexecutor.FindExecutorByType(string(catalogAction.Type))
		if actionExecutor != nil {
			err := actionExecutor.Execute(ctx, &localState, catalogAction, action)
			if err != nil {
				log.Fatal().Err(err).Str("action", action.ID).Msg("action error")
				return
			}
		} else {
			log.Error().Str("action", action.ID).Str("type", string(catalogAction.Type)).Msg("action type is not supported")
		}

		// state: store
		state.PersistStateToFile(stateFile, localState)

		// complete
		log.Info().Str("action", action.ID).Str("duration", time.Since(start).String()).Msg("action completed")
	} else {
		log.Debug().Str("action", action.ID).Msg("no workflow rule matches, not running action")
	}
}
