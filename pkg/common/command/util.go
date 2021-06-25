package command

import (
	"os"
	"strings"
)

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
