package containeraction

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"

	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/cidverseutils/network"
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
	return string(catalog.ActionTypeContainer)
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, action *catalog.WorkflowAction) error {
	// api (port or socket)
	freePort, err := network.FreePort()
	if err != nil {
		log.Fatal().Err(err).Msg("no free ports available")
	}
	apiPort := strconv.Itoa(freePort)

	// properties
	secret := generateSecret(32)
	buildID := generateSnowflakeId()
	jobID := generateSnowflakeId()

	// pass config
	var actionConfig string
	if len(ctx.Config) > 0 {
		actionConfigJSON, _ := json.Marshal(action.Config)
		actionConfig = string(actionConfigJSON)
	}

	// temp dir override
	osTempDir := os.TempDir()
	if os.Getenv("CID_TEMP_DIR") != "" {
		osTempDir = os.Getenv("CID_TEMP_DIR")
		log.Debug().Str("dir", osTempDir).Msg("overriding temp dir")
	}

	// create temp dir
	tempDir, err := os.MkdirTemp(osTempDir, "cid-job-")
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating temporary directory")
	}
	log.Debug().Str("dir", tempDir).Msg("using temp dir")
	defer func() {
		log.Debug().Str("dir", tempDir).Msg("cleaning up temp dir")
		_ = os.RemoveAll(tempDir)
	}()

	// create socket file
	socketFile := path.Join(tempDir, hash.UUIDNoDash(uuid.New().String())+".socket")

	// listen
	apiEngine := restapi.Setup(&restapi.APIConfig{
		BuildID:       buildID,
		JobID:         jobID,
		ProjectDir:    ctx.ProjectDir,
		Modules:       ctx.Modules,
		CurrentModule: ctx.CurrentModule,
		CurrentAction: catalogAction,
		Env:           ctx.Env,
		ActionConfig:  actionConfig,
		State:         localState,
		TempDir:       tempDir,
		ArtifactDir:   filepath.Join(ctx.ProjectDir, ".dist"),
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

	// configure container
	containerExec := containerruntime.Container{
		Image:            catalogAction.Container.Image,
		WorkingDirectory: ci.ToUnixPath(ctx.ProjectDir),
		Command:          insertCommandVariables(catalogAction.Container.Command, *catalogAction),
		User:             util.GetContainerUser(),
	}

	// mount project dir
	containerExec.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    ctx.ProjectDir,
		Target:    ci.ToUnixPath(ctx.ProjectDir),
	})

	// mount temp dir
	containerExec.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    tempDir,
		Target:    constants.TempPathInContainer,
	})

	if runtime.GOOS == "windows" {
		// windows does not support unix sockets
		containerExec.UserArgs = "--net host"
		containerExec.AddEnvironmentVariable("CID_API_ADDR", "http://host.docker.internal:"+apiPort)
	} else {
		// socket-based sharing of the api is more secure than sharing the host network
		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    socketFile,
			Target:    constants.SocketPathInContainer,
		})
		containerExec.AddEnvironmentVariable("CID_API_SOCKET", constants.SocketPathInContainer)
	}
	containerExec.AddEnvironmentVariable("CID_API_SECRET", secret)

	// enterprise (proxy, ca-certs)
	command.ApplyProxyConfiguration(&containerExec)
	for _, cert := range catalogAction.Container.Certs {
		certPath, certErr := util.GetCertFileByType(cert.Type)
		if certErr != nil {
			return certErr
		}

		command.ApplyCertMount(&containerExec, certPath, cert.ContainerPath)
	}

	// catalogAction access
	if len(catalogAction.Access.Env) > 0 {
		for k, v := range ctx.Env {
			for _, access := range catalogAction.Access.Env {
				if access.Pattern && regexp.MustCompile(access.Value).MatchString(k) {
					containerExec.AddEnvironmentVariable(k, v)
				} else if access.Value == k {
					containerExec.AddEnvironmentVariable(k, v)
				}
			}
		}
	}

	containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if containerCmdErr != nil {
		return containerCmdErr
	}
	log.Debug().Str("action", catalogAction.Name).Msg("container command for action: " + containerCmd)
	cmdErr := command.RunOptionalCommand(containerCmd, nil, "")
	if cmdErr != nil {
		var exitErr *exec.ExitError
		exitCode := 1
		if errors.As(cmdErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		log.Error().Int("exit_code", exitCode).Str("message", cmdErr.Error()).Msg("command failed")
		return cmdErr
	}

	return nil
}
