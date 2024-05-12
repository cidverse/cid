package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/version"
	"github.com/cidverse/normalizeci/pkg/normalizer/api"
	"github.com/rs/zerolog/log"
)

type ExecutionType string

const (
	ExecutionExec      ExecutionType = "exec"
	ExecutionContainer ExecutionType = "container"
	ExecutionNixShell  ExecutionType = "nix-shell"
)

type PreferVersion string

const (
	PreferHighest PreferVersion = "highest"
	PreferLowest  PreferVersion = "lowest"
)

type BinaryExecutionCandidate struct {
	Binary  string
	Version string
	Type    ExecutionType

	// File holds the absolute path to the executable file
	File string

	// Image holds the container image
	Image string

	// ImageCache holds information about caching for containers
	ImageCache []catalog.ImageCache

	// Mounts
	Mounts []catalog.ContainerMount

	// Security
	Security catalog.Security

	// Entrypoint overwrites the container entrypoint
	Entrypoint *string

	// Certs holds information to mount ca certificates into the containers
	Certs []catalog.ImageCerts `yaml:"certs,omitempty"`
}

// FindExecutionCandidates returns a full list of all available execution options for the specified binary
func (c *CIDConfig) FindExecutionCandidates(binary string, constraint string, preferExecutionType ExecutionType, preferVersion PreferVersion) []BinaryExecutionCandidate { //nolint:gocyclo
	var options []BinaryExecutionCandidate

	// container
	for _, entry := range c.Registry.ContainerImages {
		for _, provided := range entry.Provides {
			if binary == provided.Binary || slices.Contains(provided.Alias, binary) {
				log.Trace().Str("version", provided.Version).Str("constraint", constraint).Str("binary", binary).Str("image", entry.Image).Msg("checking version constraint")
				if version.FulfillsConstraint(provided.Version, constraint) {
					options = append(options, BinaryExecutionCandidate{
						Binary:     provided.Binary,
						Version:    provided.Version,
						Type:       ExecutionContainer,
						Image:      entry.Image,
						ImageCache: entry.Cache,
						Mounts:     entry.Mounts,
						Security:   entry.Security,
						Entrypoint: entry.Entrypoint,
						Certs:      entry.Certs,
					})
				}
			}
		}
	}

	// exec
	env := api.GetMachineEnvironment()
	for _, entry := range c.LocalTools {
		if slices.Contains(entry.Binary, binary) {
			for _, lookup := range entry.Lookup {
				// special case - PATH
				if lookup.Key == "PATH" {
					file, fileErr := exec.LookPath(binary)
					if fileErr == nil && filesystem.FileExists(file) {
						options = append(options, BinaryExecutionCandidate{
							Binary:  binary,
							Version: "0.0.0",
							Type:    ExecutionExec,
							File:    file,
						})
					} else {
						log.Warn().Str("binary", binary).Msg("didn't find binary in PATH")
					}
				}
				// check main env name
				if env[lookup.Key] != "" {
					if version.FulfillsConstraint(lookup.Version, constraint) {
						file := findExecutable(env[lookup.Key]+entry.Path, binary)
						if filesystem.FileExists(file) {
							options = append(options, BinaryExecutionCandidate{
								Binary:  binary,
								Version: lookup.Version,
								Type:    ExecutionExec,
								File:    file,
							})
						}
					}
				}
				// check with all possible suffixes
				for _, envSuffix := range entry.LookupSuffixes {
					if env[lookup.Key+envSuffix] != "" {
						if version.FulfillsConstraint(lookup.Version, constraint) {
							file := findExecutable(env[lookup.Key+envSuffix]+entry.Path, binary)
							if filesystem.FileExists(file) {
								options = append(options, BinaryExecutionCandidate{
									Binary:  binary,
									Version: lookup.Version,
									Type:    ExecutionExec,
									File:    file,
								})
							}
						}
					}
				}
			}
		}
	}

	// sort by execution type
	sort.Slice(options, func(i, j int) bool {
		// compare by executionType if different
		if options[i].Type != options[j].Type {
			if preferExecutionType == ExecutionContainer {
				return options[i].Type == ExecutionContainer
			} else if preferExecutionType == ExecutionExec {
				return options[i].Type == ExecutionExec
			}
		}

		// compare by version
		if preferVersion == PreferHighest {
			result, _ := version.Compare(options[i].Version, options[j].Version)
			return result > 0
		} else {
			result, _ := version.Compare(options[i].Version, options[j].Version)
			return result < 0
		}
	})

	for optIndex, opt := range options {
		log.Trace().Str("binary", binary).Int("index", optIndex).Interface("option", opt).Msg("identified candidate")
	}

	return options
}

// FindImageOfBinary retrieves information about the container image for the specified binary fulfilling the constraint
func (c *CIDConfig) FindImageOfBinary(binary string, constraint string) *catalog.ContainerImage {
	// lookup
	for _, entry := range c.Registry.ContainerImages {
		for _, provided := range entry.Provides {
			if binary == provided.Binary || slices.Contains(provided.Alias, binary) {
				log.Trace().Str("version", provided.Version).Str("constraint", constraint).Str("binary", binary).Str("image", entry.Image).Msg("checking version constraint")
				if version.FulfillsConstraint(provided.Version, constraint) {
					return &entry
				}
			}
		}
	}

	return nil
}

// FindPathOfBinary retrieves information about the local path of the specified binary fulfilling the constraint
func (c *CIDConfig) FindPathOfBinary(binary string, constraint string) *ToolLocal {
	// lookup
	env := api.GetMachineEnvironment()
	for _, entry := range c.LocalTools {
		if slices.Contains(entry.Binary, binary) {
			for _, lookup := range entry.Lookup {
				// special case - PATH
				if lookup.Key == "PATH" {
					file, fileErr := exec.LookPath(binary)
					if fileErr != nil {
						return nil
					}

					entry.ResolvedBinary = file
					return &entry
				}
				// check main env name
				if len(env[lookup.Key]) > 0 {
					if version.FulfillsConstraint(lookup.Version, constraint) {
						entry.ResolvedBinary = findExecutable(env[lookup.Key], binary)
						return &entry
					}
				}
				// check with all possible suffixes
				for _, envSuffix := range entry.LookupSuffixes {
					if len(env[lookup.Key+envSuffix]) > 0 {
						if version.FulfillsConstraint(lookup.Version, constraint) {
							entry.ResolvedBinary = findExecutable(env[lookup.Key+envSuffix], binary)
							return &entry
						}
					}
				}
			}
		}
	}

	return nil
}

func findExecutable(path string, file string) string {
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
			return filepath.Join(path, file)
		}
	}

	return ""
}
