package tools

import (
	"bytes"
	_ "embed"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed tools.yaml
var configContent string

const PathSeparator = string(os.PathSeparator)

type ToolCacheDir struct {
	Id string
	ContainerPath string `yaml:"dir"`
}

type ToolExecutableDiscovery struct {
	Executable string
	ExecutableFile string
	EnvironmentName string `yaml:"env-name"`
	EnvironmentNameSuffix []string `yaml:"env-allowed-suffix"`
	SubPath string `yaml:"env-path-dir"`
	Version string
}

type ToolContainerDiscovery struct {
	Executable string
	Image string
	Version string
	Cache []ToolCacheDir
}

var Config = struct {
	Tools []ToolExecutableDiscovery `yaml:"tools"`
	ContainerImages []ToolContainerDiscovery `yaml:"container-images"`
}{}

var localToolCache = make(map[string]ToolExecutableDiscovery)
var imageToolCache = make(map[string]ToolContainerDiscovery)

var toolEnvironmentDiscovery []ToolExecutableDiscovery
var toolImageDiscovery []ToolContainerDiscovery

func init() {
	// load config
	err := yaml.Unmarshal([]byte(configContent), &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load embedded tool configuration file")
	}

	toolEnvironmentDiscovery = Config.Tools
	toolImageDiscovery = Config.ContainerImages
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