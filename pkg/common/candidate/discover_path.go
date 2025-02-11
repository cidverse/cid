package candidate

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cidverse/normalizeci/pkg/normalizer/api"
	"github.com/rs/zerolog/log"
)

type PathDiscoveryRuleLookup struct {
	Key            string   // env name
	KeyAliases     []string `yaml:"key-aliases"`
	Directory      string   // directory
	Version        string   // version
	VersionCommand string   `yaml:"version-command"`
	VersionRegex   string   `yaml:"version-regex"`
}

type PathDiscoveryRule struct {
	Binary []string
	Lookup []PathDiscoveryRuleLookup
	Env    map[string]string
}

type DiscoverPathOptions struct {
	LookupRules          []PathDiscoveryRule
	VersionLookupCommand bool
}

var DefaultDiscoverPathOptions = DiscoverPathOptions{
	LookupRules: []PathDiscoveryRule{
		{
			Binary: []string{"podman"},
			Lookup: []PathDiscoveryRuleLookup{
				{
					Key:            "PATH",
					Version:        "0.0.0",
					VersionCommand: "-v",
					VersionRegex:   `(?m)podman version (\d+\.\d+\.\d+)`,
				},
			},
		},
		{
			Binary: []string{"java"},
			Lookup: []PathDiscoveryRuleLookup{
				{
					Key:            "PATH",
					Version:        "0.0.0",
					VersionCommand: "--version",
					VersionRegex:   `(?m)openjdk (\d+\.\d+\.\d+)`,
				},
				{
					Key:            "JAVA_HOME_8",
					KeyAliases:     []string{"JAVA_HOME_8_X64", "JAVA_HOME_8_X86"},
					Directory:      "bin",
					Version:        "8.0.0",
					VersionCommand: "--version",
					VersionRegex:   `(?m)openjdk (\d+\.\d+\.\d+)`,
				},
				{
					Key:            "JAVA_HOME_11",
					KeyAliases:     []string{"JAVA_HOME_11_X64", "JAVA_HOME_11_X86"},
					Directory:      "bin",
					Version:        "11.0.0",
					VersionCommand: "--version",
					VersionRegex:   `(?m)openjdk (\d+\.\d+\.\d+)`,
				},
				{
					Key:            "JAVA_HOME_17",
					KeyAliases:     []string{"JAVA_HOME_17_X64", "JAVA_HOME_17_X86"},
					Directory:      "bin",
					Version:        "17.0.0",
					VersionCommand: "--version",
					VersionRegex:   `(?m)openjdk (\d+\.\d+\.\d+)`,
				},
				{
					Key:            "JAVA_HOME_21",
					KeyAliases:     []string{"JAVA_HOME_21_X64", "JAVA_HOME_21_X86"},
					Directory:      "bin",
					Version:        "21.0.0",
					VersionCommand: "--version",
					VersionRegex:   `(?m)openjdk (\d+\.\d+\.\d+)`,
				},
			},
		},
		{
			Binary: []string{"go"},
			Lookup: []PathDiscoveryRuleLookup{
				{
					Key:            "PATH",
					Version:        "0.0.0",
					VersionCommand: "version",
					VersionRegex:   `(?m)go(\d+\.\d+\.\d+)`,
				},
				{
					Key:            "GOROOT_1_21",
					KeyAliases:     []string{"GOROOT_1_21_X64"},
					Version:        "1.21.0",
					VersionCommand: "version",
					VersionRegex:   `(?m)go(\d+\.\d+\.\d+)`,
				},
				{
					Key:            "GOROOT_1_22",
					KeyAliases:     []string{"GOROOT_1_22_X64"},
					Version:        "1.22.0",
					VersionCommand: "version",
					VersionRegex:   `(?m)go(\d+\.\d+\.\d+)`,
				},
				{
					Key:            "GOROOT_1_23",
					KeyAliases:     []string{"GOROOT_1_23_X64"},
					Version:        "1.23.0",
					VersionCommand: "version",
					VersionRegex:   `(?m)go(\d+\.\d+\.\d+)`,
				},
			},
			Env: map[string]string{
				"GOPATH":     "$HOME/go",
				"GOMODCACHE": "$HOME/go/pkg/mod",
			},
		},
	},
	VersionLookupCommand: true,
}

func DiscoverPathCandidates(opts *DiscoverPathOptions) []Candidate {
	var result []Candidate
	if opts == nil {
		opts = &DefaultDiscoverPathOptions
	}

	env := api.GetMachineEnvironment()
	for _, lr := range opts.LookupRules {

		for _, lookup := range lr.Lookup {
			// special case - PATH
			if lookup.Key == "PATH" {
				for _, binary := range lr.Binary {
					file, fileErr := exec.LookPath(binary)
					if fileErr == nil {
						if candidate := createExecCandidate(binary, file, lr.Env, lookup, opts.VersionLookupCommand); candidate != nil {
							result = append(result, *candidate)
						}
					}
				}

				continue
			}

			// check env keys
			envLookupKeys := append([]string{lookup.Key}, lookup.KeyAliases...)
			for _, envKey := range envLookupKeys {
				if env[envKey] != "" {
					for _, binary := range lr.Binary {
						file := findExecutableInDirectory(filepath.Join(env[envKey], lookup.Directory), binary)

						if candidate := createExecCandidate(binary, file, lr.Env, lookup, opts.VersionLookupCommand); candidate != nil {
							result = append(result, *candidate)
						}
					}
				}
			}
		}
	}

	return result
}

// Helper function to check if the binary exists and return a candidate
func createExecCandidate(binary, executableFile string, env map[string]string, lookup PathDiscoveryRuleLookup, versionLookupCommand bool) *ExecCandidate {
	// resolve symlink
	info, err := os.Lstat(executableFile)
	if err != nil {
		log.Error().Err(err).Str("file", executableFile).Msg("failed to get file info")
		return nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err := filepath.EvalSymlinks(executableFile)
		if err != nil {
			log.Error().Err(err).Str("file", executableFile).Msg("failed to resolve symlink")
			return nil
		}
		executableFile = resolvedPath
	}

	if _, err = os.Stat(executableFile); os.IsNotExist(err) {
		return nil
	}

	version := lookup.Version
	if lookup.VersionCommand != "" && versionLookupCommand {
		version, _ = getCommandVersion(executableFile, lookup.VersionCommand, lookup.VersionRegex)
	}

	return &ExecCandidate{
		BaseCandidate: BaseCandidate{
			Name:    binary,
			Type:    ExecutionExec,
			Version: version,
		},
		AbsolutePath: executableFile,
		Env:          env,
	}
}
