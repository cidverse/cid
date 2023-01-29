package command

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/samber/lo"

	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
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
		return "", "", err
	}

	return stdoutBuff.String(), stderrBuff.String(), nil
}

// RunAPICommand gets called from actions or the api to execute commands
func RunAPICommand(command string, env map[string]string, projectDir string, workDir string, capture bool, ports []int, userProvidedConstraint string) (stdout string, stderr string, err error) {
	var stdoutWriter io.Writer
	var stderrWriter io.Writer
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if capture {
		stdoutWriter = protectoutput.NewProtectedWriter(nil, &stdoutBuffer)
		stderrWriter = protectoutput.NewProtectedWriter(nil, &stderrBuffer)
	} else {
		stdoutWriter = protectoutput.NewProtectedWriter(os.Stdout, nil)
		stderrWriter = protectoutput.NewProtectedWriter(os.Stderr, nil)
	}

	// identify command
	args := strings.SplitN(command, " ", 2)
	binary := args[0]

	// find version constraint from config
	cmdConstraint := ">= 0.0.0"
	// constraint from config
	if value, ok := config.Current.Dependencies[binary]; ok {
		cmdConstraint = value
	}
	// user provided constraint
	if len(userProvidedConstraint) > 0 {
		cmdConstraint = userProvidedConstraint
	}

	// find execution options
	candidates := config.Current.FindExecutionCandidates(binary, cmdConstraint, config.ExecutionContainer, config.PreferHighest)
	for _, candidate := range candidates {
		// only process type ExecutionContainer
		if candidate.Type != config.ExecutionContainer {
			continue
		}

		containerExec := containerruntime.Container{}
		containerExec.SetImage(candidate.Image)
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: projectDir, Target: cihelper.ToUnixPath(projectDir)})
		containerExec.SetWorkingDirectory(cihelper.ToUnixPath(workDir))
		containerExec.SetCommand(cihelper.ToUnixPathArgs(strings.Join(args, " ")))

		// security
		if candidate.Security.Privileged {
			containerExec.SetPrivileged(true)
		}
		for _, capability := range candidate.Security.Capabilities {
			containerExec.AddCapability(capability)
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

		// ports
		for _, port := range ports {
			if IsFreePort(port) {
				containerExec.AddContainerPort(containerruntime.ContainerPort{Source: port, Target: port})
			} else {
				freePort, _ := GetFreePort()
				containerExec.AddContainerPort(containerruntime.ContainerPort{Source: freePort, Target: port})
			}
		}

		// generate and execute command
		containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
		if containerCmdErr != nil {
			return "", "", errors.New("failed to generate command: " + containerCmdErr.Error())
		}

		containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
		err := RunSystemCommand(containerCmdArgs[0], containerCmdArgs[1], env, "", stdoutWriter, stderrWriter)
		if err != nil {
			return "", "", errors.New("command failed: " + err.Error())
		}

		return strings.TrimSuffix(stdoutBuffer.String(), "\r\n"), strings.TrimSuffix(stderrBuffer.String(), "\r\n"), nil
	}

	return "", "", errors.New("no method to execute command: " + binary)
}

func runCommand(command string, env map[string]string, projectDir string, workDir string, stdout io.Writer, stderr io.Writer) error {
	cmdArgs := strings.SplitN(command, " ", 2)
	originalBinary := cmdArgs[0]
	cmdPayload := ""
	if len(cmdArgs) == 2 {
		cmdPayload = cmdArgs[1]
	}

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
		return RunSystemCommand(candidate.File, cmdPayload, env, workDir, stdout, stderr)
	case config.ExecutionContainer:
		containerExec := containerruntime.Container{}

		if projectDir == "" {
			projectDir = vcsrepository.FindRepositoryDirectory(workDir)
		}

		containerExec.SetImage(candidate.Image)
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: projectDir, Target: cihelper.ToUnixPath(projectDir)})
		containerExec.SetWorkingDirectory(cihelper.ToUnixPath(workDir))
		containerExec.SetCommand(cihelper.ToUnixPathArgs(strings.Join(cmdArgs, " ")))

		// security
		if candidate.Security.Privileged {
			containerExec.SetPrivileged(true)
		}
		for _, capability := range candidate.Security.Capabilities {
			containerExec.AddCapability(capability)
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
		return RunSystemCommand(containerCmdArgs[0], containerCmdArgs[1], env, workDir, stdout, stderr)
	default:
		log.Fatal().Interface("type", candidate.Type).Msg("execution type is not supported!")
	}

	return nil
}

// RunSystemCommand runs a command and forwards all output to current console session
func RunSystemCommand(file string, args string, env map[string]string, workDir string, stdout io.Writer, stderr io.Writer) error {
	log.Trace().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("command exec")

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
