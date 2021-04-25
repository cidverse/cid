package tools

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const PathSeparator = string(os.PathSeparator)

type ToolExecutableDiscovery struct {
	Executable string
	EnvironmentName string
	SubPath string
	Version string
}

// FindLocalTool tries to find a tool/cli fulfilling the specified version constraints in the local environment
func FindLocalTool(executable string, constraint string) (string, error) {
	var toolEnvironmentDiscovery []ToolExecutableDiscovery
	// golang
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_16", SubPath: "/bin", Version: "1.16.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_15", SubPath: "/bin", Version: "1.15.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_14", SubPath: "/bin", Version: "1.14.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_13", SubPath: "/bin", Version: "1.13.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_12", SubPath: "/bin", Version: "1.12.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_11", SubPath: "/bin", Version: "1.11.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_10", SubPath: "/bin", Version: "1.10.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_16", SubPath: "/bin", Version: "1.16.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_15", SubPath: "/bin", Version: "1.15.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_14", SubPath: "/bin", Version: "1.14.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_13", SubPath: "/bin", Version: "1.13.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_12", SubPath: "/bin", Version: "1.12.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_11", SubPath: "/bin", Version: "1.11.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_10", SubPath: "/bin", Version: "1.10.0"})
	// java
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_17", SubPath: "/bin", Version: "17.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_16", SubPath: "/bin", Version: "16.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_15", SubPath: "/bin", Version: "15.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_14", SubPath: "/bin", Version: "14.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_13", SubPath: "/bin", Version: "13.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_12", SubPath: "/bin", Version: "12.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_11", SubPath: "/bin", Version: "11.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_10", SubPath: "/bin", Version: "10.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_9", SubPath: "/bin", Version: "9.0.0"})
	toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_8", SubPath: "/bin", Version: "8.0.0"})

	// check based on env paths
	systemEnvironment := os.Environ()
	for _, env := range systemEnvironment {
		z := strings.SplitN(env, "=", 2)
		key := z[0]
		value := z[1]

		for _, entry := range toolEnvironmentDiscovery {
			if executable == entry.Executable {
				if strings.Contains(key, entry.EnvironmentName) && IsVersionFulfillingConstraint(entry.Version, constraint) {
					return FindExecutable(value+entry.SubPath, entry.Executable), nil
				}
			}
		}
	}

	return "", errors.New("failed to find executable")
}

func IsVersionFulfillingConstraint(version string, constraint string) bool {
	// constraint
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		log.Debug().Err(err).Str("constraint", version).Msg("version constraint is unparsable")
		return false
	}

	// version
	v, err := semver.NewVersion(version)
	if err != nil {
		log.Debug().Err(err).Str("version", version).Msg("version is unparsable")
		return false
	}

	// check
	ok, validateErr := c.Validate(v)
	if ok == false {
		var allErrors bytes.Buffer

		for _, err := range validateErr {
			allErrors.WriteString(err.Error())
		}

		log.Debug().Str("version", version).Str("constraint", constraint).Str("error", allErrors.String()).Msg("version does not fulfill constraint")
		return false
	}

	return true
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
			return path+PathSeparator+file
		}
	}

	return ""
}