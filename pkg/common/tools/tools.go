package tools

import (
	"bytes"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
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
	ExecutableFile string
	EnvironmentName string
	EnvironmentNameSuffix []string
	SubPath string
	Version string
}

type ToolContainerDiscovery struct {
	Executable string
	Image string
	Version string
	Cache []ToolCacheDir
}

var localToolCache = make(map[string]ToolExecutableDiscovery)
var imageToolCache = make(map[string]ToolContainerDiscovery)

var toolEnvironmentDiscovery []ToolExecutableDiscovery
var toolImageDiscovery []ToolContainerDiscovery

func init() {
	// init tool lookup
	// golang
	for _, element := range []string{"16", "15", "14", "13", "12", "11", "10"} {
		toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "go", EnvironmentName: "GOROOT_1_"+element, EnvironmentNameSuffix: []string{"_X64"}, SubPath: "/bin", Version: "1."+element+".0"})
		toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "gofmt", EnvironmentName: "GOROOT_1_"+element, EnvironmentNameSuffix: []string{"_X64"}, SubPath: "/bin", Version: "1."+element+".0"})
	}
	// java
	for _, element := range []string{"17", "16", "15", "14", "13", "12", "11", "10", "9", "8"} {
		toolEnvironmentDiscovery = append(toolEnvironmentDiscovery, ToolExecutableDiscovery{Executable: "java", EnvironmentName: "JAVA_HOME_"+element, EnvironmentNameSuffix: []string{"_X64"}, SubPath: "/bin", Version: element+".0.0"})
	}

	// init image lookup
	// golang
	for _, element := range []string{"1.16.4", "1.16.3", "1.16.2", "1.16.1", "1.16.0", "1.15.12", "1.15.11", "1.15.10", "1.15.9", "1.15.8", "1.15.7", "1.15.6", "1.15.5", "1.15.4", "1.15.3", "1.15.2", "1.15.1", "1.15.0"} {
		toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "go", Image: "golang:"+element+"-alpine", Version: element, Cache: []ToolCacheDir{{"go-pkg", "/go/pkg"}}})
		toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "gofmt", Image: "golang:"+element+"-alpine", Version: element, Cache: []ToolCacheDir{{"go-pkg", "/go/pkg"}}})
	}
	// golangci-lint
	toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "golangci-lint", Image: "golangci/golangci-lint:v1.40.1-alpine", Version: "1.40.1"})
	// java
	toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "java", Image: "adoptopenjdk/openjdk16:jdk-16.0.1_9", Version: "16.0.1"})
	toolImageDiscovery = append(toolImageDiscovery, ToolContainerDiscovery{Executable: "java", Image: "adoptopenjdk/openjdk15:jdk-15.0.2_7", Version: "15.0.2"})
	// upx
}

// FindLocalTool tries to find a tool/cli fulfilling the specified version constraints in the local environment
func FindLocalTool(executable string, constraint string) (ToolExecutableDiscovery, error) {
	// get from cache
	if funk.Contains(localToolCache, executable+"/"+constraint) {
		return localToolCache[executable+"/"+constraint], nil
	}

	// check based on env paths
	env := common.GetMachineEnvironment()
	for _, entry := range toolEnvironmentDiscovery {
		if executable == entry.Executable {
			// check main env name
			if len(env[entry.EnvironmentName]) > 0 {
				if IsVersionFulfillingConstraint(entry.Version, constraint) {
					entry.ExecutableFile = FindExecutable(env[entry.EnvironmentName]+entry.SubPath, entry.Executable)
					localToolCache[executable+"/"+constraint] = entry
					return entry, nil
				}
			}
			// check with all possible suffixes
			for _, envSuffix := range entry.EnvironmentNameSuffix {
				if len(env[entry.EnvironmentName+envSuffix]) > 0 {
					if IsVersionFulfillingConstraint(entry.Version, constraint) {
						entry.ExecutableFile = FindExecutable(env[entry.EnvironmentName+envSuffix]+entry.SubPath, entry.Executable)
						localToolCache[executable+"/"+constraint] = entry
						return entry, nil
					}
				}
			}
		}
	}

	return ToolExecutableDiscovery{}, errors.New("failed to find executable")
}

func FindContainerImage(executable string, constraint string) (ToolContainerDiscovery, error) {
	// get from cache
	if funk.Contains(imageToolCache, executable+"/"+constraint) {
		return imageToolCache[executable+"/"+constraint], nil
	}

	// check based on env paths
	for _, entry := range toolImageDiscovery {
		if executable == entry.Executable {
			if IsVersionFulfillingConstraint(entry.Version, constraint) {
				imageToolCache[executable+"/"+constraint] = entry
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
	if !ok {
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