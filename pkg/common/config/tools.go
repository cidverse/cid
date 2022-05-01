package config

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/version"
	"github.com/cidverse/normalizeci/pkg/common"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const PathSeparator = string(os.PathSeparator)

type ToolCacheDir struct {
	Id            string
	ContainerPath string `yaml:"dir"`
	MountType     string `yaml:"type"`
}

type ToolExecutableDiscovery struct {
	Executable            string
	ExecutableFile        string
	EnvironmentName       string   `yaml:"env-name"`
	EnvironmentNameSuffix []string `yaml:"env-allowed-suffix"`
	SubPath               string   `yaml:"env-path-dir"`
	Version               string
}

type ToolContainerDiscovery struct {
	Executable string
	Image      string
	Version    string
	Cache      []ToolCacheDir
}

// FindLocalTool tries to find a tool/cli fulfilling the specified version constraints in the local environment
func FindLocalTool(executable string, constraint string) (ToolExecutableDiscovery, error) {
	// check based on env paths
	env := common.GetMachineEnvironment()
	for _, entry := range Config.Tools {
		if executable == entry.Executable {
			// special case - PATH
			if entry.EnvironmentName == "PATH" {
				file, fileErr := exec.LookPath(executable)
				if fileErr != nil {
					return ToolExecutableDiscovery{}, fileErr
				}

				entry.ExecutableFile = file
				return entry, nil
			}
			// check main env name
			if len(env[entry.EnvironmentName]) > 0 {
				if version.FulfillsConstraint(entry.Version, constraint) {
					entry.ExecutableFile = FindExecutable(env[entry.EnvironmentName]+entry.SubPath, entry.Executable)
					return entry, nil
				}
			}
			// check with all possible suffixes
			for _, envSuffix := range entry.EnvironmentNameSuffix {
				if len(env[entry.EnvironmentName+envSuffix]) > 0 {
					if version.FulfillsConstraint(entry.Version, constraint) {
						entry.ExecutableFile = FindExecutable(env[entry.EnvironmentName+envSuffix]+entry.SubPath, entry.Executable)
						return entry, nil
					}
				}
			}
		}
	}

	return ToolExecutableDiscovery{}, errors.New("failed to find executable")
}

func FindContainerImage(executable string, constraint string) (ToolContainerDiscovery, error) {
	// check based on env paths
	for _, entry := range Config.ContainerImages {
		if executable == entry.Executable {
			if version.FulfillsConstraint(entry.Version, constraint) {
				return entry, nil
			}
		}
	}
	return ToolContainerDiscovery{}, errors.New("failed to find image")
}

func FindExecutable(path string, file string) string {
	if runtime.GOOS == "windows" {
		// windows
		if _, err := os.Stat(filepath.Join(path, file+".exe")); err == nil {
			return filepath.Join(path, file+".exe")
		}
		if _, err := os.Stat(filepath.Join(path, file+".bat")); err == nil {
			return filepath.Join(path, file+".bat")
		}
		if _, err := os.Stat(filepath.Join(path, file+".ps1")); err == nil {
			return filepath.Join(path, file+".ps1")
		}
	} else {
		// unix
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return path + PathSeparator + file
		}
	}

	return ""
}
