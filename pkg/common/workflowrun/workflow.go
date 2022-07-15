package workflowrun

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
	"time"
)

// IsWorkflowExecutable returns true if the workflow is executable (enabled + at least one rule matches)
func IsWorkflowExecutable(w config.Workflow, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(w.Rules, rules.GetRuleContext(env))

	if w.Enabled == true && (len(w.Rules) == 0 || matchingRules > 0) {
		return true
	}

	return false
}

// IsStageExecutable returns true if the stage is executable (enabled + at least one rule matches)
func IsStageExecutable(s config.WorkflowStage, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(s.Rules, rules.GetRuleContext(env))

	if s.Enabled == true && (len(s.Rules) == 0 || matchingRules > 0) {
		return true
	}

	return false
}

// IsActionExecutable returns true if the action is executable (enabled + at least one rule matches)
func IsActionExecutable(a config.Action, env map[string]string) bool {
	matchingRules := rules.EvaluateRules(a.Rules, rules.GetRuleContext(env))

	if a.Enabled == true && (len(a.Rules) == 0 || matchingRules > 0) {
		return true
	}

	return false
}

// FirstWorkflowMatchingRules returns the first workflow that matches at least one rule
func FirstWorkflowMatchingRules(workflows []config.Workflow, env map[string]string) *config.Workflow {
	// select workflow
	log.Info().Msg("evaluating all workflows")
	for _, wf := range workflows {
		if wf.Enabled != true {
			log.Debug().Str("workflow", wf.Name).Msg("workflow is disabled, skipping")
			continue
		}
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

func RunWorkflow(cfg *config.CIDConfig, wf *config.Workflow, env map[string]string, projectDir string, stagesFilter []string, modulesFilter []string) {
	log.Debug().Str("workflow", wf.Name).Msg("workflow start")
	start := time.Now()
	ruleContext := rules.GetRuleContext(env)

	if rules.AnyRuleMatches(wf.Rules, ruleContext) {
		for _, stage := range wf.Stages {
			if len(stagesFilter) == 0 || funk.Contains(stagesFilter, stage.Name) {
				RunWorkflowStage(cfg, &stage, env, projectDir, modulesFilter)
			} else {
				log.Debug().Str("workflow", wf.Name).Str("stage", stage.Name).Strs("filter", stagesFilter).Msg("stage has been skipped")
			}
		}

		// complete
		log.Info().Str("workflow", wf.Name).Str("duration", time.Since(start).String()).Msg("workflow completed")
	} else {
		log.Debug().Str("workflow", wf.Name).Msg("no workflow rule matches, not running workflow")
	}
}

func RunWorkflowStage(cfg *config.CIDConfig, stage *config.WorkflowStage, env map[string]string, projectDir string, modulesFilter []string) {
	log.Debug().Str("stage", stage.Name).Msg("stage start")
	start := time.Now()
	ruleContext := rules.GetRuleContext(env)

	if rules.AnyRuleMatches(stage.Rules, ruleContext) {
		for _, action := range stage.Actions {
			RunWorkflowAction(cfg, &action, env, projectDir, modulesFilter)
		}

		// complete
		log.Info().Str("stage", stage.Name).Str("duration", time.Since(start).String()).Msg("stage completed")
	} else {
		log.Debug().Str("stage", stage.Name).Msg("no workflow rule matches, not running stage")
	}
}

func RunWorkflowAction(cfg *config.CIDConfig, action *config.WorkflowAction, env map[string]string, projectDir string, modulesFilter []string) {
	log.Debug().Str("action", action.Id).Msg("action start")
	catalogAction := cfg.FindAction(action.Id)
	ctx := api.GetActionContext(projectDir, env, nil)

	// serialize action config for pass-thru
	configAsYaml, _ := yaml.Marshal(&action.Config)
	ctx.Config = string(configAsYaml)

	// project-scoped actions
	if catalogAction.Scope == config.ActionScopeProject {
		runWorkflowAction(catalogAction, action, ctx)
	}

	// module-scoped actions
	if catalogAction.Scope == config.ActionScopeModule {
		// for each module
		for _, module := range ctx.Modules {
			moduleRef := *module

			// customize context
			ctx.CurrentModule = &moduleRef
			api.UpdateContext(&ctx)

			// check module filter
			if len(modulesFilter) > 0 && !funk.Contains(modulesFilter, moduleRef.Name) {
				continue
			}

			runWorkflowAction(catalogAction, action, ctx)
		}
	}
}

func runWorkflowAction(catalogAction *config.Action, action *config.WorkflowAction, ctx api.ActionExecutionContext) {
	start := time.Now()
	ruleContext := rules.GetRuleContext(ctx.Env)
	if rules.AnyRuleMatches(action.Rules, ruleContext) {
		// state: retrieve/init
		state := getState(ctx)

		// serialize action config for pass-thru
		actConfig, _ := yaml.Marshal(&action.Config)
		log.Trace().Str("action", action.Id).Str("type", string(catalogAction.Type)).Str("config", string(actConfig)).Msg("action configuration")

		// execute
		if catalogAction.Type == config.ActionTypeBuiltinGolang {
			if evaluateActionBuiltinGolang(&ctx, &state, catalogAction, action) {
				runActionBuiltinGolang(&ctx, &state, catalogAction, action)
			} else {
				log.Debug().Str("action", action.Id).Msg("requirements not fulfilled, not running action")
				return
			}
		} else {
			log.Error().Str("action", action.Id).Str("type", string(catalogAction.Type)).Msg("action type is not supported")
		}

		// state: store
		persistState(ctx, state)

		// complete
		log.Info().Str("action", action.Id).Str("duration", time.Since(start).String()).Msg("action completed")
	} else {
		log.Debug().Str("action", action.Id).Msg("no workflow rule matches, not running action")
	}
}
