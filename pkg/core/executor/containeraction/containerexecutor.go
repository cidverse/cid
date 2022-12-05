package containeraction

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	_ "github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type Executor struct{}

func (e Executor) GetName() string {
	return "container"
}

func (e Executor) GetVersion() string {
	return "0.1.0"
}

func (e Executor) GetType() string {
	return string(config.ActionTypeContainer)
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) error {
	// properties
	apiPort := strconv.Itoa(findAvailablePort())
	socketFile := path.Join(ctx.Paths.Temp, "api-"+uuid.New().String()+".socket")
	secret := generateSecret()

	// pass config
	var actionConfig string
	if len(ctx.Config) > 0 {
		actionConfigJSON, _ := json.Marshal(action.Config)
		actionConfig = string(actionConfigJSON)
	}

	// listen
	apiEngine := restapi.Setup(restapi.APIConfig{
		ProjectDir:    ctx.ProjectDir,
		Modules:       ctx.Modules,
		CurrentModule: ctx.CurrentModule,
		CurrentAction: catalogAction,
		Env:           ctx.Env,
		ActionConfig:  actionConfig,
	})
	restapi.SecureWithAPIKey(apiEngine, secret)
	go func() {
		if runtime.GOOS == "windows" {
			restapi.ListenOnAddr(apiEngine, ":"+apiPort)
		} else {
			restapi.ListenOnSocket(apiEngine, socketFile)
		}
	}()

	// shutdown listener (on function end)
	defer func(apiEngine *echo.Echo, ctx context.Context) {
		err := apiEngine.Shutdown(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to shutdown rest api")
		}
	}(apiEngine, context.Background())

	defer func() {
		if _, err := os.Stat(socketFile); err == nil {
			_ = os.Remove(socketFile)
		}
	}()

	// wait a short moment for the unix socket to be created / the api endpoint to be ready
	time.Sleep(100 * time.Millisecond)

	// configure container
	containerExec := containerruntime.Container{}
	containerExec.SetImage(catalogAction.Container.Image)
	containerExec.SetCommand(insertCommandVariables(catalogAction.Container.Command, *catalogAction))
	containerExec.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    ctx.ProjectDir,
		Target:    cihelper.ToUnixPath(ctx.ProjectDir),
	})
	containerExec.SetWorkingDirectory(cihelper.ToUnixPath(ctx.ProjectDir))

	if runtime.GOOS == "windows" {
		// windows does not support unix sockets
		containerExec.SetUserArgs("--net host")
		containerExec.AddEnvironmentVariable("CID_API_ADDR", "http://host.docker.internal:"+apiPort)
	} else {
		// socket-based sharing of the api is more secure than sharing the host network
		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    socketFile,
			Target:    socketFile,
		})
		containerExec.AddEnvironmentVariable("CID_API_SOCKET", socketFile)
	}
	containerExec.AddEnvironmentVariable("CID_API_SECRET", secret)

	containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if containerCmdErr != nil {
		return containerCmdErr
	}

	log.Debug().Str("action", catalogAction.Name).Msg("container command for action: " + containerCmd)
	_, err := command.RunCommandAndGetOutput(containerCmd, nil, "")
	return err
}
