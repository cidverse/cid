package tools

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"runtime"
)

const PathSeparator = string(os.PathSeparator)

type ToolCacheDir struct {
	Id string
	ContainerPath string
}

type ToolExecutableDiscovery struct {
	Executable string
	EnvironmentName string
	SubPath string
	Version string
}

type ToolContainerDiscovery struct {
	Executable string
	Image string
	Version string
	Cache []ToolCacheDir
}

// FindLocalTool tries to find a tool/cli fulfilling the specified version constraints in the local environment
func FindLocalTool(executable string, constraint string) (ToolExecutableDiscovery, string, error) {
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
	env := common.GetMachineEnvironment()
	for _, entry := range toolEnvironmentDiscovery {
		if executable == entry.Executable && len(env[entry.EnvironmentName]) > 0 {
			if IsVersionFulfillingConstraint(entry.Version, constraint) {
				return entry, FindExecutable(env[entry.EnvironmentName]+entry.SubPath, entry.Executable), nil
			}
		}
	}

	return ToolExecutableDiscovery{}, "", errors.New("failed to find executable")
}

func FindContainerImage(executable string, constraint string) (ToolContainerDiscovery, error) {
	var toolImageDiscovery []ToolContainerDiscovery
	// golang
	for _, element := range []string{"1.16.4", "1.16.3", "1.16.2", "1.16.1", "1.16.0", "1.15.12", "1.15.11", "1.15.10", "1.15.9", "1.15.8", "1.15.7", "1.15.6", "1.15.5", "1.15.4", "1.15.3", "1.15.2", "1.15.1", "1.15.0"} {
		toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "go", Image: "golang:"+element+"-alpine", Version: element, Cache: []ToolCacheDir{{"go-pkg", "/go/pkg"}}})
		toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "gofmt", Image: "golang:"+element+"-alpine", Version: element, Cache: []ToolCacheDir{{"go-pkg", "/go/pkg"}}})
	}
	// golangci-lint
	toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "golangci-lint", Image: "golangci/golangci-lint:v1.40.1-alpine", Version: "1.40.1"})

	// check based on env paths
	for _, entry := range toolImageDiscovery {
		if executable == entry.Executable {
			if IsVersionFulfillingConstraint(entry.Version, constraint) {
				return entry, nil
			}
		}
	}

	return ToolContainerDiscovery{}, errors.New("failed to find image")
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