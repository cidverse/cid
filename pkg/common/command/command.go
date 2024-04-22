package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/network"
	"github.com/cidverse/go-vcs/vcsutil"
	"github.com/samber/lo"

	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/rs/zerolog/log"
)

const AnyVersionConstraint = ">= 0.0.0"

// GetCommandVersion returns the version of an executable
func GetCommandVersion(binary string) (string, error) {
	// find version constraint from config
	binaryVersionConstraint := AnyVersionConstraint
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
	err := runCommand(command, env, "", workDir, protectoutput.NewProtectedWriter(os.Stdout, nil), protectoutput.NewProtectedWriter(os.Stderr, nil))
	if err != nil {
		log.Fatal().Err(err).Str("command", command).Msg("failed to execute command")
	}
}

// RunOptionalCommand runs a command and forwards all output to console
func RunOptionalCommand(command string, env map[string]string, workDir string) error {
	return runCommand(command, env, "", workDir, protectoutput.NewProtectedWriter(os.Stdout, nil), protectoutput.NewProtectedWriter(os.Stderr, nil))
}

// RunCommandAndGetOutput runs a command and returns the full response / command output
func RunCommandAndGetOutput(command string, env map[string]string, workDir string) (string, string, error) {
	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer

	err := runCommand(command, env, "", workDir, &stdoutBuff, &stderrBuff)
	if err != nil {
		return stdoutBuff.String(), stderrBuff.String(), err
	}

	return stdoutBuff.String(), stderrBuff.String(), nil
}

type APICommandExecute struct {
	Command                string
	Env                    map[string]string
	ProjectDir             string
	WorkDir                string
	TempDir                string
	Capture                bool
	Ports                  []int
	UserProvidedConstraint string
	Stdin                  io.Reader
}

// RunAPICommand gets called from actions or the api to execute commands
func RunAPICommand(cmd APICommandExecute) (stdout string, stderr string, executionCandidate *config.BinaryExecutionCandidate, err error) {
	var stdoutWriter io.Writer
	var stderrWriter io.Writer
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if cmd.Capture {
		stdoutWriter = protectoutput.NewProtectedWriter(nil, &stdoutBuffer)
		stderrWriter = protectoutput.NewProtectedWriter(nil, &stderrBuffer)
	} else {
		stdoutWriter = protectoutput.NewProtectedWriter(os.Stdout, nil)
		stderrWriter = protectoutput.NewProtectedWriter(os.Stderr, nil)
	}

	// identify command
	args := strings.SplitN(cmd.Command, " ", 2)
	binary := args[0]

	// find version constraint from config
	cmdConstraint := AnyVersionConstraint
	// constraint from config
	if value, ok := config.Current.Dependencies[binary]; ok {
		cmdConstraint = value
	}
	// user provided constraint
	if len(cmd.UserProvidedConstraint) > 0 {
		cmdConstraint = cmd.UserProvidedConstraint
	}

	// find execution options
	candidates := config.Current.FindExecutionCandidates(binary, cmdConstraint, config.ExecutionContainer, config.PreferHighest)
	for _, candidate := range candidates {
		// only process type ExecutionContainer
		if candidate.Type != config.ExecutionContainer {
			continue
		}

		// overwrite binary for alias use-case
		args[0] = candidate.Binary

		containerExec := containerruntime.Container{
			Image:            candidate.Image,
			WorkingDirectory: ci.ToUnixPath(cmd.WorkDir),
			Entrypoint:       candidate.Entrypoint,
			Command:          ci.ToUnixPathArgs(strings.Join(args, " ")),
			User:             util.GetContainerUser(),
		}

		// mount project dir
		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    cmd.ProjectDir,
			Target:    ci.ToUnixPath(cmd.ProjectDir),
		})

		// mount temp dir
		if cmd.TempDir != "" {
			containerExec.AddVolume(containerruntime.ContainerMount{
				MountType: "directory",
				Source:    cmd.TempDir,
				Target:    constants.TempPathInContainer,
			})
		}

		// interactive?
		if cmd.Stdin != nil {
			containerExec.Interactive = true
			containerExec.TTY = true
		}

		// security
		if candidate.Security.Privileged {
			containerExec.Privileged = true
		}
		containerExec.Capabilities = append(containerExec.Capabilities, candidate.Security.Capabilities...)

		// mounts
		for _, mount := range candidate.Mounts {
			containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: mount.Src, Target: mount.Dest})
		}

		// add env + sort by key
		sortedEnvKeys := lo.Keys(cmd.Env)
		sort.Strings(sortedEnvKeys)
		for _, key := range sortedEnvKeys {
			containerExec.AddEnvironmentVariable(key, cmd.Env[key])
		}

		// cache
		for _, c := range candidate.ImageCache {
			dir := filepath.Join(util.GetUserConfigDirectory(), "cid-cache-"+c.ID)
			_ = os.MkdirAll(dir, 0775)
			containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: dir, Target: c.ContainerPath})
		}

		// ports
		for _, port := range cmd.Ports {
			if network.IsFreePort(port) {
				containerExec.ContainerPorts = append(containerExec.ContainerPorts, containerruntime.ContainerPort{Source: port, Target: port})
			} else {
				freePort, _ := network.FreePort()
				containerExec.ContainerPorts = append(containerExec.ContainerPorts, containerruntime.ContainerPort{Source: freePort, Target: port})
			}
		}

		// enterprise (proxy, ca-certs)
		ApplyProxyConfiguration(&containerExec)
		for _, cert := range candidate.Certs {
			ApplyCertMount(&containerExec, GetCertFileByType(cert.Type), cert.ContainerPath)
		}

		// generate and execute command
		containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
		if containerCmdErr != nil {
			return "", "", &candidate, errors.New("failed to generate command: " + containerCmdErr.Error())
		}
		log.Debug().Msg("running command via api: " + containerCmd)

		containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
		err := RunSystemCommand(containerCmdArgs[0], containerCmdArgs[1], cmd.Env, "", cmd.Stdin, stdoutWriter, stderrWriter)
		if err != nil {
			return "", "", &candidate, errors.New("command failed: " + err.Error())
		}

		return strings.TrimSuffix(stdoutBuffer.String(), "\r\n"), strings.TrimSuffix(stderrBuffer.String(), "\r\n"), &candidate, nil
	}

	return "", "", nil, errors.New("no method to execute command: " + binary)
}

func runCommand(command string, env map[string]string, projectDir string, workDir string, stdout io.Writer, stderr io.Writer) error {
	cmdArgs := strings.SplitN(command, " ", 2)
	originalBinary := cmdArgs[0]
	cmdPayload := ""
	if len(cmdArgs) == 2 {
		cmdPayload = cmdArgs[1]
	}

	// find version constraint from config
	cmdConstraint := AnyVersionConstraint
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
		return RunSystemCommand(candidate.File, cmdPayload, env, workDir, nil, stdout, stderr)
	case config.ExecutionContainer:
		if projectDir == "" {
			p, err := vcsutil.FindProjectDirectory(workDir)
			if err != nil {
				return fmt.Errorf("failed to find project directory: %s", err.Error())
			}
			projectDir = p
		}

		// overwrite binary for alias use-case
		cmdArgs[0] = candidate.Binary

		containerExec := containerruntime.Container{
			Image:            candidate.Image,
			WorkingDirectory: ci.ToUnixPath(workDir),
			Entrypoint:       candidate.Entrypoint,
			Command:          ci.ToUnixPathArgs(strings.Join(cmdArgs, " ")),
			User:             util.GetContainerUser(),
		}
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: projectDir, Target: ci.ToUnixPath(projectDir)})

		// security
		if candidate.Security.Privileged {
			containerExec.Privileged = true
		}
		for _, capability := range candidate.Security.Capabilities {
			containerExec.Capabilities = append(containerExec.Capabilities, capability)
		}

		// mounts
		for _, mount := range candidate.Mounts {
			containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: mount.Src, Target: mount.Dest})
		}

		// add env + sort by key
		sortedEnvKeys := lo.Keys(env)
		sort.Strings(sortedEnvKeys)
		for _, key := range sortedEnvKeys {
			containerExec.AddEnvironmentVariable(key, env[key])
		}

		// cache
		for _, c := range candidate.ImageCache {
			containerExec.AddVolume(containerruntime.ContainerMount{MountType: "volume", Source: "cid-cache-" + c.ID, Target: c.ContainerPath})
		}

		containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
		if containerCmdErr != nil {
			return containerCmdErr
		}

		log.Debug().Msg("container-exec: " + containerCmd)
		containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
		return RunSystemCommand(containerCmdArgs[0], containerCmdArgs[1], env, workDir, nil, stdout, stderr)
	default:
		log.Fatal().Interface("type", candidate.Type).Msg("execution type is not supported!")
	}

	return nil
}

// RunSystemCommand runs a command and forwards all output to current console session
func RunSystemCommand(file string, args string, env map[string]string, workDir string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	log.Trace().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("command exec")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, file, args, workDir)
	if cmdErr != nil {
		log.Err(cmdErr).Msg("failed to execute command")
		return cmdErr
	}

	var commandEnv = make(map[string]string)
	for _, line := range os.Environ() {
		// TODO: maybe only PATH is enough? commandEnv["PATH"] = os.Getenv("PATH")
		z := strings.SplitN(line, "=", 2)
		commandEnv[z[0]] = z[1]
	}
	for k, v := range env {
		commandEnv[k] = v
	}
	cmd.Env = ci.EnvMapToStringSlice(commandEnv)
	cmd.Dir = workDir
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		log.Trace().Err(err).Str("command_result", "error").Msg(file + " " + args)
		return err
	}

	log.Trace().Str("command_result", "ok").Msg(file + " " + args)
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
