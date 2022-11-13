package containeraction

import (
	"context"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/core/state"
	_ "github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
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
	socketFile := path.Join(ctx.Paths.Temp, "my.socket")
	secret := generateSecret()

	// listen
	apiEngine := restapi.Setup(ctx.ProjectDir, ctx.Modules, ctx.CurrentModule, ctx.Env)
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

	// configure container
	containerExec := containerruntime.Container{}
	containerExec.SetImage(catalogAction.Container.Image)
	containerExec.SetCommand(insertCommandVariables(catalogAction.Container.Command, *catalogAction))

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

	containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
	return command.RunSystemCommandPassThru(containerCmdArgs[0], containerCmdArgs[1], nil, "", os.Stdout, os.Stderr)
}
