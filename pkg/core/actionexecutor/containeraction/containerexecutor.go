package containeraction

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/cidverse/cid/internal/state"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/actionexecutor/api"
	"github.com/cidverse/cid/pkg/core/actionexecutor/builtin"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/cidverseutils/network"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/google/uuid"
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

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, step plangenerate.Step) error {
	// api (port or socket)
	freePort, err := network.FreePort()
	if err != nil {
		return fmt.Errorf("could not get free port: %w", err)
	}
	apiPort := strconv.Itoa(freePort)

	// temp dir
	tempBaseDir, err := util.CITempDir(ctx.NCI.ServiceSlug)
	if err != nil {
		return err
	}

	// properties
	secret := api.GenerateSecret(32)
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
		SDKClient: builtin.ActionSDK{
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
		},
	})
	restapi.SecureWithAPIKey(apiEngine, secret)
	go func() {
		if runtime.GOOS == "windows" {
			restapi.ListenOnAddr(apiEngine, ":"+apiPort)
		} else {
			restapi.ListenOnSocket(apiEngine, socketFile)
		}
	}()

	// wait a short moment for the unix socket to be created / the api endpoint to be ready
	time.Sleep(100 * time.Millisecond) // TODO: find a better way to wait for the api to be ready

	// configure container
	containerExec := containerruntime.Container{
		Image:            catalogAction.Container.Image,
		WorkingDirectory: ci.ToUnixPath(ctx.ProjectDir),
		Command:          api.InsertCommandVariables(catalogAction.Container.Command, *catalogAction),
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
	if len(catalogAction.Metadata.Access.Environment) > 0 {
		for k, v := range ctx.Env {
			for _, access := range catalogAction.Metadata.Access.Environment {
				if access.Pattern && regexp.MustCompile(access.Name).MatchString(k) {
					containerExec.AddEnvironmentVariable(k, v)
				} else if access.Name == k {
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
		exitCode := 1

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}

		slog.With("err", err).With("exit_code", exitCode).Error("command failed")
		return err
	}

	return nil
}
