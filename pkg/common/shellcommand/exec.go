package shellcommand

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cidverse/cidverseutils/ci"
	"github.com/rs/zerolog/log"
)

func Command(command string) *exec.Cmd {
	args, _ := SplitCommand(command)

	return exec.Command(args[0], args[1:]...)
}

func SplitCommand(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for i := 0; i < len(command); i++ {
		c := command[i]
		if escapeNext {
			current.WriteByte(c)
			escapeNext = false
			continue
		}
		if c == '\\' {
			escapeNext = true
			continue
		}
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(c)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if inQuotes {
		return nil, fmt.Errorf("unmatched quotes in command string")
	}

	return args, nil
}

// PrepareCommand prepares a command to be executed
func PrepareCommand(command string, platform string, shell string, fullEnv bool, env map[string]string, workDir string, stdin io.Reader, stdoutWriter io.Writer, stderrWriter io.Writer) (*exec.Cmd, error) {
	args, err := FormatPlatformCommand(command, platform, shell)
	if err != nil {
		return nil, err
	}
	cmd := Command(args)
	cmd.Dir = workDir
	cmd.Stdin = stdin
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	log.Trace().Str("command", command).Str("platform", platform).Str("shell", shell).Str("workdir", workDir).Strs("args", cmd.Args).Interface("env", env).Msg("preparing command")

	var commandEnv = make(map[string]string)
	if fullEnv {
		for _, line := range os.Environ() {
			z := strings.SplitN(line, "=", 2)
			commandEnv[z[0]] = z[1]
		}
	} else {
		commandEnv["PATH"] = os.Getenv("PATH")
		commandEnv["HOME"] = os.Getenv("HOME")
	}
	for k, v := range env {
		commandEnv[k] = v
	}
	cmd.Env = ci.EnvMapToStringSlice(commandEnv)

	return cmd, nil
}
