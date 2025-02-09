package candidate

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func findExecutableInDirectory(dir string, file string) string {
	extensions := []string{""}
	if runtime.GOOS == "windows" {
		extensions = []string{".exe", ".bat", ".ps1"}
	}

	for _, ext := range extensions {
		executablePath := filepath.Join(dir, file+ext)
		info, err := os.Stat(executablePath)
		if err == nil {
			// require executable bit on Unix
			if runtime.GOOS != "windows" && (info.Mode()&0111) == 0 {
				continue
			}

			return executablePath
		}
	}

	return ""
}

func getCommandVersion(command, args, regex string) (string, error) {
	cmd := exec.Command(command, strings.Split(args, " ")...)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Join(err, errors.New("failed to execute command"))
	}

	output := strings.TrimSpace(string(out))
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to compile regex for version extraction"))
	}
	log.Debug().Str("command", command+" "+args).Str("regex", regex).Str("output", output).Msg("finding version via command")

	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", errors.New("failed to extract version from command output")
}

func findExecutablesInDirectory(dir string) []string {
	var executables []string

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("failed to read directory")
		return executables
	}

	for _, file := range files {
		if file.Type()&os.ModeSymlink != 0 {
			resolvedPath, err := filepath.EvalSymlinks(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Error().Err(err).Str("file", file.Name()).Msg("failed to resolve symlink")
				continue
			}

			if info, err := os.Stat(resolvedPath); err == nil && info.Mode().IsRegular() {
				executables = append(executables, file.Name())
			}
		} else if file.Type().IsRegular() {
			executables = append(executables, file.Name())
		}
	}

	return executables
}

// ReplaceCommandPlaceholders replaces env placeholders in a command
func ReplaceCommandPlaceholders(input string, env map[string]string) string {
	// timestamp
	input = strings.ReplaceAll(input, "{TIMESTAMP_RFC3339}", time.Now().Format(time.RFC3339))

	// env
	for k, v := range env {
		input = strings.ReplaceAll(input, "{"+k+"}", v)
	}

	return input
}
