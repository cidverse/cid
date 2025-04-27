package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/builtin/builtinaction"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/actionexecutor/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/cidverseutils/network"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
	// api (port or socket)
	freePort, err := network.FreePort()
	if err != nil {
		log.Fatal().Err(err).Msg("no free ports available")
	}
	apiPort := strconv.Itoa(freePort)

	// temp dir override
	osTempDir := os.TempDir()
	if os.Getenv("CID_TEMP_DIR") != "" {
		osTempDir = os.Getenv("CID_TEMP_DIR")
		log.Debug().Str("dir", osTempDir).Msg("overriding temp dir")
	}

	// properties
	secret := api.GenerateSecret(32)
	buildID := api.GenerateSnowflakeId()
	jobID := api.GenerateSnowflakeId()
	artifactDir := filepath.Join(ctx.ProjectDir, ".dist")
	tempDir, err := os.MkdirTemp(osTempDir, "cid-job-")
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

	// create socket file
	socketFile := path.Join(tempDir, hash.UUIDNoDash(uuid.New().String())+".socket")

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

	// listen
	apiEngine := restapi.Setup(&restapi.APIConfig{
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
	})
	restapi.SecureWithAPIKey(apiEngine, secret)
	go func() {
		if runtime.GOOS == "windows" {
			restapi.ListenOnAddr(apiEngine, ":"+apiPort)
		} else {
			restapi.ListenOnSocket(apiEngine, socketFile)
		}
	}()

	// shutdown listener
	defer func(apiEngine *echo.Echo, ctx context.Context) {
		err = apiEngine.Shutdown(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to shutdown rest api")
		}
	}(apiEngine, context.Background())

	// wait a short moment for the unix socket to be created / the api endpoint to be ready
	time.Sleep(100 * time.Millisecond) // TODO: find a better way to wait for the api to be ready

	// run action
	sdkConfig := cidsdk.SDKConfig{}
	if runtime.GOOS == "windows" {
		sdkConfig.APIAddr = "http://host.docker.internal:" + apiPort
	} else {
		sdkConfig.APISocket = socketFile
	}
	sdkConfig.APISecret = secret
	sdk, err := cidsdk.NewSDKWithConfig(sdkConfig)
	if err != nil {
		return fmt.Errorf("failed to create sdk: %w", err)
	}

	// lookup in action by name map - TODO: make function in actions for lookup
	actionLookup := builtinaction.GetActions(sdk)
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
