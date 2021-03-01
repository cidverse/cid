package command

import (
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"runtime"
)

func RunCommand(command string, env []string) error {
	workDir := filesystem.GetWorkingDirectory()
	log.Debug().Str("command", command).Str("os", runtime.GOOS).Str("workdir", workDir).Msg("Running Command")

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
