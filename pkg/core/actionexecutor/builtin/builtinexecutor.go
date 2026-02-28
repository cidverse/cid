package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/builtin/builtinaction"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/actionexecutor/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/util"
	"github.com/rs/zerolog/log"
)

type Executor struct{}

func (e Executor) GetName() string {
	return "builtin"
}

func (e Executor) GetVersion() string {
	return "0.1.0"
}

func (e Executor) GetType() string {
	return string(catalog.ActionTypeBuiltIn)
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, step plangenerate.Step) error {
	// temp dir
	tempBaseDir, err := util.CITempDir(ctx.NCI.ServiceSlug)
	if err != nil {
		return err
	}

	// properties
	buildID := api.GenerateSnowflakeId()
	jobID := api.GenerateSnowflakeId()
	artifactDir := filepath.Join(ctx.ProjectDir, ".dist")
	tempDir, err := os.MkdirTemp(tempBaseDir, "cid-job-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// create dirs
	log.Debug().Str("temp-dir", tempDir).Str("artifact-dir", artifactDir).Msg("creating action directories")
	err = os.MkdirAll(artifactDir, 0770)
	if err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}
	err = os.MkdirAll(tempDir, 0770)
	if err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}
	_ = os.Chmod(tempDir, 0770) // MkdirTemp creates the dir chmod 0700, which is not accessible for other users in the group
	defer func() {
		log.Debug().Str("dir", tempDir).Msg("cleaning up temp dir")
		_ = os.RemoveAll(tempDir)
	}()

	// executables
	executableCandidates, err := command.CandidatesFromConfig(config.Current)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to discover candidates")
	}

	// actionConfig
	actionConfig, err := json.Marshal(ctx.Config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal action config")
	}

	// sdk client
	sdkClient := ActionSDK{
		BuildID:              buildID,
		JobID:                jobID,
		ProjectDir:           ctx.ProjectDir,
		Modules:              ctx.Modules,
		Step:                 step,
		CurrentModule:        ctx.CurrentModule,
		CurrentAction:        catalogAction,
		NCI:                  ctx.NCI,
		Env:                  ctx.Env,
		ActionEnv:            ctx.ActionEnv,
		ActionConfig:         string(actionConfig),
		State:                localState,
		TempDir:              tempDir,
		ArtifactDir:          artifactDir,
		ExecutableCandidates: executableCandidates,
	}

	// lookup in action by name map - TODO: make function in actions for lookup
	actionLookup := builtinaction.GetActions(sdkClient)
	action, ok := actionLookup[catalogAction.Metadata.Name]
	if !ok {
		return fmt.Errorf("action %s not found", catalogAction.Metadata.Name)
	}

	// run action
	err = action.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute action: %w", err)
	}

	return nil
}
