package workflowrun

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/core/actionexecutor"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/repoanalyzer/analyzer"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func RunWorkflowAction(cfg *config.CIDConfig, action *catalog.WorkflowAction, env map[string]string, projectDir string, modulesFilter []string) {
	log.Debug().Str("action", action.ID).Msg("action start")
	catalogAction := cfg.Registry.FindAction(action.ID)
	if catalogAction == nil {
		log.Fatal().Str("action_id", action.ID).Msg("workflow configuration error, referencing actions that do not exist")
		os.Exit(1)
	}
	modules := analyzer.ScanDirectory(filesystem.WorkingDirOrPanic())
	ctx := api.GetActionContext(modules, projectDir, env, catalogAction.Metadata.Access)

	// serialize action config for pass-thru
	configAsJSON, _ := json.Marshal(&action.Config)
	ctx.Config = string(configAsJSON)

	// project-scoped actions
	if catalogAction.Metadata.Scope == catalog.ActionScopeProject {
		ruleContext := rules.GetProjectRuleContext(ctx.Env, ctx.Modules)
		ruleMatch := rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext)
		log.Debug().Str("Trace", action.ID).Bool("rules_match", ruleMatch).Msg("check action rules")
		if ruleMatch {
			runWorkflowAction(catalogAction, action, &ctx)
		}
	}

	// module-scoped actions
	if catalogAction.Metadata.Scope == catalog.ActionScopeModule {
		// for each module
		for _, m := range ctx.Modules {
			moduleRef := *m
			log.Trace().Str("action", action.ID).Str("module", moduleRef.Slug).Msg("action for module")

			// customize context
			ctx.CurrentModule = &moduleRef

			// check module filter
			if len(modulesFilter) > 0 && !slices.Contains(modulesFilter, moduleRef.Name) {
				log.Trace().Str("action", action.ID).Str("module", moduleRef.Slug).Strs("filter", modulesFilter).Msg("action skipped by module filter")
				continue
			}

			var ruleContext = rules.GetModuleRuleContext(ctx.Env, &moduleRef)
			ruleMatch := rules.AnyRuleMatches(append(action.Rules, catalogAction.Metadata.Rules...), ruleContext)
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
		currentModule := "root"
		if ctx.CurrentModule != nil {
			currentModule = ctx.CurrentModule.Slug
		}
		log.Info().Str("action", action.ID).Str("module", currentModule).Msg("action start")

		// state: retrieve/init
		localState := state.GetStateFromDirectory(ctx.Paths.Artifact)
		localState.Modules = ctx.Modules

		// add action to log
		localState.AuditLog = append(localState.AuditLog, state.AuditEvents{
			Timestamp: time.Now(),
			Type:      "action",
			Payload: map[string]string{
				"action": catalogAction.Repository + "/" + catalogAction.Metadata.Name + "@" + catalogAction.Version,
				"uri":    fmt.Sprintf("oci://%s", catalogAction.Container.Image),
			},
		})

		// serialize action config for pass-thru
		actConfig, _ := yaml.Marshal(&action.Config)
		log.Trace().Str("action", action.ID).Str("type", string(catalogAction.Type)).Str("config", string(actConfig)).Msg("action configuration")

		// paths
		_ = os.MkdirAll(ctx.Paths.Temp, os.ModePerm)
		_ = os.MkdirAll(ctx.Paths.Artifact, os.ModePerm)

		// execute
		actionExecutor := actionexecutor.FindExecutorByType(string(catalogAction.Type))
		if actionExecutor != nil {
			err := actionExecutor.Execute(ctx, &localState, catalogAction)
			if err != nil {
				slog.With("err", err).With("action", action.ID).Error("action execution failed")
				os.Exit(1)
				return
			}
		} else {
			log.Error().Str("action", action.ID).Str("type", string(catalogAction.Type)).Msg("action type is not supported")
		}

		// state: store
		stateFile := filepath.Join(ctx.Paths.Artifact, "state.json")
		if !strings.HasPrefix(ctx.Env["NCI_SERVICE_SLUG"], "local") {
			stateFile = filepath.Join(ctx.Paths.Artifact, fmt.Sprintf("state-%s.json", ctx.Env["NCI_PIPELINE_JOB_ID"]))
		}
		state.PersistStateToFile(stateFile, localState)

		// complete
		log.Info().Str("action", action.ID).Str("duration", time.Since(start).String()).Str("module", currentModule).Msg("action completed")
	} else {
		log.Debug().Str("action", action.ID).Msg("no workflow rule matches, not running action")
	}
}
