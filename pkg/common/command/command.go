package command

import (
	"bytes"
	"errors"
	"github.com/samber/lo"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/container_runtime"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/rs/zerolog/log"
)

// GetCommandVersion returns the version of an executable
func GetCommandVersion(binary string) (string, error) {
	// find version constraint from config
	binaryVersionConstraint := ">= 0.0.0"
	if value, ok := config.Current.Dependencies[binary]; ok {
		binaryVersionConstraint = value
	}

	// prefer local tools if we find some that match the project version constraints
	toolData := config.Current.FindPathOfBinary(binary, binaryVersionConstraint)
	if toolData != nil {
		// TODO: return toolData.Version, nil
		return "0.0.0", nil
	}

	// find container image
	containerImage := config.Current.FindImageOfBinary(binary, binaryVersionConstraint)
	if containerImage != nil {
		for _, provides := range containerImage.Provides {
			if binary == provides.Binary {
				return provides.Version, nil
			}
		}
	}

	// can't run cmd
	return "", errors.New("can't determinate version of " + binary)
}

// RunCommand runs a required command and forwards all output to console, but will panic/exit if the command fails
func RunCommand(command string, env map[string]string, workDir string) {
	err := runCommand(command, env, workDir, protectoutput.NewProtectedWriter(os.Stdout, nil), protectoutput.NewProtectedWriter(os.Stderr, nil))
	if err != nil {
		log.Fatal().Err(err).Str("command", command).Msg("failed to execute command")
	}
}

// RunOptionalCommand runs a command and forwards all output to console
func RunOptionalCommand(command string, env map[string]string, workDir string) error {
	return runCommand(command, env, workDir, protectoutput.NewProtectedWriter(os.Stdout, nil), protectoutput.NewProtectedWriter(os.Stderr, nil))
}

// RunCommandAndGetOutput runs a command and returns the full response / command output
func RunCommandAndGetOutput(command string, env map[string]string, workDir string) (string, error) {
	var resultBuff bytes.Buffer

	err := runCommand(command, env, workDir, &resultBuff, &resultBuff)
	if err != nil {
		return "", err
	}

	return resultBuff.String(), nil
}

func runCommand(command string, env map[string]string, workDir string, stdout io.Writer, stderr io.Writer) error {
	cmdArgs := strings.SplitN(command, " ", 2)
	originalBinary := cmdArgs[0]
	cmdPayload := cmdArgs[1]

	// find version constraint from config
	cmdConstraint := ">= 0.0.0"
	if value, ok := config.Current.Dependencies[originalBinary]; ok {
		cmdConstraint = value
	}

	// lookup execution options
	candidates := config.Current.FindExecutionCandidates(originalBinary, cmdConstraint, config.ExecutionExec, config.PreferHighest)
	log.Trace().Interface("candidates", candidates).Str("binary", originalBinary).Msg("identified all available execution candidates")

	// no ways to execute command
	if len(candidates) == 0 {
		return errors.New("no method available to execute command " + originalBinary)
	}

	candidate := candidates[0]
	switch candidate.Type {
	case config.ExecutionExec:
		return RunSystemCommandPassThru(candidate.File, cmdPayload, env, workDir, stdout, stderr)
	case config.ExecutionContainer:
		containerExec := container_runtime.Container{}

		projectDir := vcsrepository.FindRepositoryDirectory(workDir)

		containerExec.SetImage(candidate.Image)
		containerExec.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: cihelper.ToUnixPath(projectDir), Target: cihelper.ToUnixPath(projectDir)})
		containerExec.SetWorkingDirectory(cihelper.ToUnixPath(workDir))
		containerExec.SetEntrypoint("unset")
		containerExec.SetCommand(strings.Join(cmdArgs, " "))

		// security
		for _, cap := range candidate.Security.Capabilities {
			containerExec.AddCapability(cap)
		}

		// add env + sort by key
		sortedEnvKeys := lo.Keys(env)
		sort.Strings(sortedEnvKeys)
		for _, key := range sortedEnvKeys {
			containerExec.AddEnvironmentVariable(key, env[key])
		}

		// cache
		for _, c := range candidate.ImageCache {
			cacheDir := path.Join(os.TempDir(), "cid", c.ID)
			_ = os.MkdirAll(cacheDir, 0777)
			_ = os.Chmod(cacheDir, 0777)

			// support mounting volumes (auto created if not present) or directories
			if c.MountType == "volume" {
				containerExec.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: c.ID, Target: c.ContainerPath})
			} else {
				containerExec.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: cihelper.ToUnixPath(cacheDir), Target: c.ContainerPath})
			}
		}

		containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
		if containerCmdErr != nil {
			return containerCmdErr
		}

		log.Debug().Msg("container-exec: " + cihelper.ToUnixPathArgs(containerCmd))
		containerCmdArgs := strings.SplitN(cihelper.ToUnixPathArgs(containerCmd), " ", 2)
		return RunSystemCommandPassThru(containerCmdArgs[0], containerCmdArgs[1], env, workDir, stdout, stderr)
	default:
		log.Fatal().Interface("type", candidate.Type).Msg("execution type is not supported!")
	}

	return nil
}

// RunSystemCommandPassThru runs a command and forwards all output to current console session
func RunSystemCommandPassThru(file string, args string, env map[string]string, workDir string, stdout io.Writer, stderr io.Writer) error {
	log.Debug().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("command exec")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, file, args, workDir)
	if cmdErr != nil {
		log.Err(cmdErr).Msg("failed to execute command")
		return cmdErr
	}

	cmd.Env = getFullEnvFromMap(env)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		log.Debug().Err(err).Str("file", file).Str("args", args).Msg("command execution failed")
		return err
	}

	log.Debug().Str("file", file).Str("args", args).Msg("command execution OK")
	return nil
}

// GetPlatformSpecificCommand returns a platform-specific exec.Cmd
func GetPlatformSpecificCommand(platform string, file string, args string, workDir string) (*exec.Cmd, error) {
	if platform == "linux" {
		return exec.Command("sh", "-c", file+` `+args), nil //nolint:gosec
	} else if platform == "windows" {
		// Notes:
		// powershell needs .\ prefix for executables in the current directory
		if filesystem.FileExists(file) {
			return exec.Command("powershell", `& '`+file+`' `+args), nil //nolint:gosec
		} else if filesystem.FileExists(workDir + `/` + file + `.bat`) {
			return exec.Command("powershell", `.\`+file+`.bat `+args), nil //nolint:gosec
		} else if filesystem.FileExists(workDir + `/` + file + `.ps1`) {
			return exec.Command("powershell", `.\`+file+`.ps1 `+args), nil //nolint:gosec
		} else if filesystem.FileExists(workDir + `/` + file) {
			return exec.Command("powershell", `.\`+file+` `+args), nil //nolint:gosec
		}

		return exec.Command("powershell", file+` `+args), nil //nolint:gosec
	} else if platform == "darwin" {
		return exec.Command("sh", "-c", file+` `+args), nil //nolint:gosec
	}

	return nil, errors.New("command.GetPlatformSpecificCommand failed, platform " + platform + " is not supported!")
}
