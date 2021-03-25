package command

import (
	"bytes"
	"errors"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RunSilentCommand runs a command and stores sthe response in a string
func RunSilentCommand(command string, env []string) (string, error) {
	var resultBuff bytes.Buffer
	workDir := filesystem.GetWorkingDirectory()
	cmdArgs := strings.SplitN(command, " ", 2)
	cmdBinary := cmdArgs[0]
	log.Debug().Str("command", command).Str("binary", cmdBinary).Str("os", runtime.GOOS).Str("workdir", workDir).Msg("running command")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, command)
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
		log.Fatal().Str("command", command).Str("error", err.Error()).Msg("Command Execution Failed")
		return "", err
	}

	log.Debug().Str("command", command).Msg("Command Execution OK")
	return resultBuff.String(), nil
}

// RunCommand runs a command and forwards all output to console
func RunCommand(command string, env []string) error {
	workDir := filesystem.GetWorkingDirectory()
	cmdArgs := strings.SplitN(command, " ", 2)
	cmdBinary := cmdArgs[0]
	log.Debug().Str("command", command).Str("binary", cmdBinary).Str("os", runtime.GOOS).Str("workdir", workDir).Msg("running command")

	// Run Command
	cmd, cmdErr := GetPlatformSpecificCommand(runtime.GOOS, command)
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
		log.Fatal().Str("command", command).Str("error", err.Error()).Msg("Command Execution Failed")
		return err
	}

	log.Debug().Str("command", command).Msg("Command Execution OK")
	return nil
}

// GetPlatformSpecificCommand returns a platform-specific exec.Cmd
func GetPlatformSpecificCommand(platform string, command string) (*exec.Cmd, error) {
	workDir := filesystem.GetWorkingDirectory()
	cmdArgs := strings.SplitN(command, " ", 2)
	cmdBinary := cmdArgs[0]

	if platform == "linux" {
		return exec.Command("sh", "-c", command), nil
	} else if platform == "windows" {
		// powershell needs .\ prefix for executables in the current directory
		if _, err := os.Stat(workDir+`/`+cmdBinary+`.bat`); !os.IsNotExist(err) {
			command = `.\`+command
		} else if _, err := os.Stat(workDir+`/`+cmdBinary); !os.IsNotExist(err) {
			command = `.\`+command
		}

		return exec.Command("powershell", command), nil
	} else if platform == "darwin" {
		return exec.Command("sh", "-c", command), nil
	}

	return nil, errors.New("command.GetPlatformSpecificCommand - platform " + platform + " is not supported!")
}