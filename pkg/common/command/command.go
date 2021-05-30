package command

import (
	"bytes"
	"errors"
	"github.com/EnvCLI/EnvCLI/pkg/container_runtime"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/cidverse/x/pkg/common/config"
	"github.com/cidverse/x/pkg/common/tools"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// GetCommandVersion returns the version of a executable
func GetCommandVersion(executable string) (string, error) {
	// find version constraint from config
	cmdConstraint := ">= 0.0.0"
	if value, ok := config.Config.Dependencies["bin/"+executable]; ok {
		cmdConstraint = value
	}

	// prefer local tools if we find some that match the project version constraints
	toolData, toolErr := tools.FindLocalTool(executable, cmdConstraint)
	if toolErr == nil {
		return toolData.Version, nil
	}

	// find container image
	containerImage, containerImageErr := tools.FindContainerImage(executable, cmdConstraint)
	if containerImageErr == nil {
		return containerImage.Version, nil
	}

	// can't run cmd
	return "", errors.New("can't determinate version of " + executable)
}

// RunCommand runs a command and forwards all output to console
func RunCommand(command string, env map[string]string, workDir string) error {
	cmdArgs := strings.SplitN(command, " ", 2)
	originalBinary := cmdArgs[0]
	cmdPayload := cmdArgs[1]

	// find version constraint from config
	cmdConstraint := ">= 0.0.0"
	if value, ok := config.Config.Dependencies["bin/"+originalBinary]; ok {
		cmdConstraint = value
	}

	// local execution
	cmdBinary := ""
	localTool, localToolErr := tools.FindLocalTool(originalBinary, cmdConstraint)
	if localToolErr == nil {
		cmdBinary = localTool.ExecutableFile
	}

	// container execution
	containerImage, containerImageErr := tools.FindContainerImage(originalBinary, cmdConstraint)
	containerExec := container_runtime.Container{}
	if containerImageErr == nil {
		projectDir := vcsrepository.FindRepositoryDirectory(workDir)

		containerExec.SetImage(containerImage.Image)
		containerExec.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: cihelper.ToUnixPath(projectDir), Target: cihelper.ToUnixPath(projectDir)})
		containerExec.SetWorkingDirectory(cihelper.ToUnixPath(workDir))
		containerExec.SetEntrypoint("unset")
		containerExec.SetCommand(strings.Join(cmdArgs, " "))
		for key, value := range env {
			containerExec.AddEnvironmentVariable(key, value)
		}

		// cache
		for _, c := range containerImage.Cache {
			cacheDir := path.Join(os.TempDir(), "cid", c.Id)
			_ = os.MkdirAll(cacheDir, os.ModePerm)

			containerExec.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: cihelper.ToUnixPath(cacheDir), Target: c.ContainerPath})
		}
	}

	// decide how to execute this command
	log.Debug().Str("executable", originalBinary).Str("args", cmdPayload).Str("os", runtime.GOOS).Str("workdir", workDir).Str("mode", string(config.Config.Mode)).Str("localpath", cmdBinary).Msg("command info")
	if config.Config.Mode == config.PreferLocal {
		if len(cmdBinary) > 0 {
			// run locally
			return RunSystemCommandPassThru(cmdBinary, cmdPayload, env, workDir)
		} else if containerImageErr == nil && len(containerImage.Image) > 0 {
			// run in container
			containerCmd := cihelper.ToUnixPathArgs(containerExec.GetRunCommand(containerExec.DetectRuntime()))
			log.Debug().Msg("container-exec: " + containerCmd)
			containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
			return RunSystemCommandPassThru(containerCmdArgs[0], containerCmdArgs[1], env, workDir)
		} else {
			log.Fatal().Str("executable", originalBinary).Msg("no method available to execute command")
		}
	} else if config.Config.Mode == config.Strict {
		if containerImageErr == nil && len(containerImage.Image) > 0 {
			// run in container
			containerCmd := cihelper.ToUnixPathArgs(containerExec.GetRunCommand(containerExec.DetectRuntime()))
			log.Debug().Msg("container-exec: " + containerCmd)
			containerCmdArgs := strings.SplitN(containerCmd, " ", 2)
			return RunSystemCommandPassThru(containerCmdArgs[0], containerCmdArgs[1], env, workDir)
		} else if len(cmdBinary) > 0 {
			// run locally
			return RunSystemCommandPassThru(cmdBinary, cmdPayload, env, workDir)
		} else {
			log.Fatal().Str("executable", originalBinary).Msg("no method available to execute command")
		}
	} else {
		log.Fatal().Str("mode", string(config.Config.Mode)).Msg("execution mode not supported")
	}

	// can't run cmd
	log.Fatal().Str("executable", originalBinary).Msg("no method available to execute command")
	return errors.New("no method available to execute command " + originalBinary)
}

// RunSystemCommand runs a command and stores the response in a string
func RunSystemCommand(file string, args string, env map[string]string, workDir string) (string, error) {
	var resultBuff bytes.Buffer
	log.Debug().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("running command")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, file, args, workDir)
	if cmdErr != nil {
		log.Err(cmdErr).Msg("failed to execute command")
		return "", cmdErr
	}

	cmd.Env = getFullEnvFromMap(env)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = &resultBuff
	cmd.Stderr = &resultBuff
	err := cmd.Run()
	if err != nil {
		log.Fatal().Str("file", file).Str("args", args).Str("error", err.Error()).Msg("Command Execution Failed")
		return "", err
	}

	log.Debug().Str("file", file).Str("args", args).Msg("Command Execution OK")
	return resultBuff.String(), nil
}

// RunSystemCommandPassThru runs a command and forwards all output to current console session
func RunSystemCommandPassThru(file string, args string, env map[string]string, workDir string) error {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal().Str("file", file).Str("args", args).Str("error", err.Error()).Msg("command execution failed")
		return err
	}

	log.Debug().Str("file", file).Str("args", args).Msg("command execution OK")
	return nil
}

// GetPlatformSpecificCommand returns a platform-specific exec.Cmd
func GetPlatformSpecificCommand(platform string, file string, args string, workDir string) (*exec.Cmd, error) {
	if platform == "linux" {
		return exec.Command("sh", "-c", file+` `+args), nil
	} else if platform == "windows" {
		// Notes:
		// powershell needs .\ prefix for executables in the current directory
		if filesystem.FileExists(file) {
			return exec.Command("powershell", `& '`+file+`' `+args), nil
		} else if filesystem.FileExists(workDir+`/`+file+`.bat`) {
			return exec.Command("powershell", `.\`+file+`.bat `+args), nil
		} else if filesystem.FileExists(workDir+`/`+file+`.ps1`) {
			return exec.Command("powershell", `.\`+file+`.ps1 `+args), nil
		} else if filesystem.FileExists(workDir+`/`+file) {
			return exec.Command("powershell", `.\`+file+` `+args), nil
		}

		return exec.Command("powershell", file+` `+args), nil
	} else if platform == "darwin" {
		return exec.Command("sh", "-c", file+` `+args), nil
	}

	return nil, errors.New("command.GetPlatformSpecificCommand failed, platform " + platform + " is not supported!")
}

func getFullEnvFromMap(env map[string]string) []string {
	// full environment
	fullEnv := make(map[string]string)
	for _, line := range os.Environ() {
		z := strings.SplitN(line, "=", 2)
		fullEnv[z[0]] = z[1]
	}
	// custom env parameters
	for k, v := range env {
		fullEnv[k] = v
	}

	// turn into a slice
	var envLines []string
	for k, v := range fullEnv {
		envLines = append(envLines, k+"="+v)
	}

	return envLines
}
