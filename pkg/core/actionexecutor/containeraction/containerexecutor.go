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
	"sync"
	"time"

	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/cidverseutils/network"
	"github.com/cidverse/cidverseutils/redact"
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

	// executables
	executableCandidates, err := executable.LoadExecutables()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load candidates from cache")
		os.Exit(1)
	}

	// listen
	apiEngine := restapi.Setup(&restapi.APIConfig{
		BuildID:              buildID,
		JobID:                jobID,
		ProjectDir:           ctx.ProjectDir,
		Modules:              ctx.Modules,
		CurrentModule:        ctx.CurrentModule,
		CurrentAction:        catalogAction,
		Env:                  ctx.Env,
		ActionConfig:         actionConfig,
		State:                localState,
		TempDir:              tempDir,
		ArtifactDir:          filepath.Join(ctx.ProjectDir, ".dist"),
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
		Target:    tempDir,
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
	containerExec.AutoProxyConfiguration()
	for _, cert := range catalogAction.Container.Certs {
		certPath, certErr := util.GetCertFileByType(cert.Type)
		if certErr != nil {
			return certErr
		}

		// copy files into a custom directory if CID_CERT_MOUNT_DIR is set, workaround for some dind setups
		customCertDir := os.Getenv("CID_CERT_MOUNT_DIR")
		if customCertDir != "" {
			_ = os.MkdirAll(customCertDir, os.ModePerm)
			certDestinationFile := filepath.Join(customCertDir, filepath.Base(certPath))
			_ = filesystem.CopyFile(certPath, certDestinationFile)

			certPath = certDestinationFile
		}

		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    certPath,
			Target:    cert.ContainerPath,
			Mode:      containerruntime.ReadMode,
		})
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

	containerCmd, err := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if err != nil {
		return err
	}

	cmd, err := shellcommand.PrepareCommand(containerCmd, runtime.GOOS, "", true, nil, "", nil, redact.NewProtectedWriter(os.Stdout, nil, &sync.Mutex{}, nil), redact.NewProtectedWriter(os.Stderr, nil, &sync.Mutex{}, nil))
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		exitCode := 1
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		log.Error().Int("exit_code", exitCode).Str("message", err.Error()).Msg("command failed")
		return err
	}

	return nil
}
