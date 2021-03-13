package api

import (
	"os"
	"strings"
)

// Normalizer is a common interface to work with all normalizers
type ActionStep interface {
	GetStage() string
	GetName() string
	Check(projectDir string) bool
	Execute(projectDir string, env []string, args []string)
}

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact string `default:"dist"`
	Cache string `default:""`
}

// GetValueFromEnv gets a env value from a list of environment variables
func GetValueFromEnv(env []string, key string) string {
	for _, envvar := range env {
		z := strings.SplitN(envvar, "=", 2)
		if strings.ToLower(key) == strings.ToLower(z[0]) {
			return strings.ToLower(z[1])
		}
	}

	return ""
}

// GetEffectiveEnv returns the effective environment
func GetEffectiveEnv(env []string) []string {
	// Environment
	finalEnv := os.Environ()
	finalEnv = append(finalEnv, env...)

	return finalEnv
}

// GetCacheDir returns the caching directory for a module
func GetCacheDir(pathConfig PathConfig, module string) string {
	if len(pathConfig.Cache) > 0 {
		return pathConfig.Cache + `/` + module
	}

	return os.TempDir() + `/.cid/` + module
}