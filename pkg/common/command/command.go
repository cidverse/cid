package command

import (
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func RunCommand(command string, env []string) error {
	workDir := filesystem.GetWorkingDirectory()
	cmdArgs := strings.SplitN(command, " ", 2)
	cmdBinary := cmdArgs[0]
	log.Debug().Str("command", command).Str("binary", cmdBinary).Str("os", runtime.GOOS).Str("workdir", workDir).Msg("Running Command")

	// Run Command
	if runtime.GOOS == "linux" {
		cmd := exec.Command("/usr/bin/env", "sh", "-c", command)
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
	} else if runtime.GOOS == "windows" {
		// powershell needs .\ prefix for executables in the current directory
		if _, err := os.Stat(workDir+cmdBinary); !os.IsNotExist(err) {
			command = `.\`+command
		}

		cmd := exec.Command("powershell", command)
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
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", command)
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
	}

	return nil
}
