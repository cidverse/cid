package command

import (
	"bytes"
	"errors"
	"github.com/PhilippHeuer/cid/pkg/common/config"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/PhilippHeuer/cid/pkg/common/tools"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RunCommand runs a command and forwards all output to console
func RunCommand(command string, env []string, workDir string) error {
	cmdArgs := strings.SplitN(command, " ", 2)
	originalBinary := cmdArgs[0]
	cmdBinary := originalBinary
	cmdPayload := cmdArgs[1]

	// decide how to execute this command
	if config.Config.Mode == config.PreferLocal {
		// find version constraint from config
		cmdConstraint := ">= 0.0.0"
		if value, ok := config.Config.Dependencies["bin/"+originalBinary]; ok {
			cmdConstraint = value
		}

		// prefer local tools if we find some that match the project version constraints
		tool, toolErr := tools.FindLocalTool(cmdBinary, cmdConstraint)
		if toolErr == nil {
			cmdBinary = tool
		}
	}

	// TODO: try to find container image

	log.Debug().Str("commandBinary", originalBinary).Str("commandPayload", cmdPayload).Str("os", runtime.GOOS).Str("workdir", workDir).Msg("running command")
	err := RunSystemCommandPassThru(cmdBinary, cmdPayload, env, workDir)

	return err
}

// RunSystemCommand runs a command and stores the response in a string
func RunSystemCommand(file string, args string, env []string, workDir string) (string, error) {
	var resultBuff bytes.Buffer
	log.Debug().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("running command")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, file, args, workDir)
	if cmdErr != nil {
		log.Err(cmdErr).Msg("failed to execute command")
		return "", cmdErr
	}

	cmd.Env = env
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
func RunSystemCommandPassThru(file string, args string, env []string, workDir string) error {
	log.Debug().Str("file", file).Str("args", args).Str("workdir", workDir).Msg("running command")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, file, args, workDir)
	if cmdErr != nil {
		log.Err(cmdErr).Msg("failed to execute command")
		return cmdErr
	}

	cmd.Env = env
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal().Str("file", file).Str("args", args).Str("error", err.Error()).Msg("Command Execution Failed")
		return err
	}

	log.Debug().Str("file", file).Str("args", args).Msg("Command Execution OK")
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
