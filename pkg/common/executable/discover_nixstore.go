package executable

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/util"
	"github.com/rs/zerolog/log"
)

type NixStorePackage struct {
	Name             string
	ProgramsProvided []string
	Expression       string
	Env              map[string]string
}

type DiscoverNixStoreOptions struct {
	Packages             []NixStorePackage
	VersionLookupCommand bool
}

var DefaultDiscoverNixOptions = DiscoverNixStoreOptions{
	Packages: []NixStorePackage{
		{
			Name:       "openjdk",
			Expression: `([a-z0-9]{32})-(openjdk)-(\d+\.\d+\.\d+.+)`,
			Env: map[string]string{
				"JAVA_HOME": "{{EXECUTABLE_PKG_DIR}}/lib/openjdk",
			},
		},
		{
			Name:       "go",
			Expression: `([a-z0-9]{32})-(go)-(\d+\.\d+\.\d+)`,
			Env: map[string]string{
				"GOPATH":     "$HOME/go",
				"GOMODCACHE": "$HOME/go/pkg/mod",
				"GOCACHE":    "$HOME/.cache/go-build",
			},
		},
		{
			Name:       "zig",
			Expression: `([a-z0-9]{32})-(zig)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:       "helm",
			Expression: `([a-z0-9]{32})-(kubernetes-helm)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:       "helmfile",
			Expression: `([a-z0-9]{32})-(helmfile)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:       "kubectl",
			Expression: `([a-z0-9]{32})-(kubectl)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:             "openshift",
			ProgramsProvided: []string{"oc"},
			Expression:       `([a-z0-9]{32})-(openshift)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:       "ansible",
			Expression: `([a-z0-9]{32})-(ansible)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:             "minio-client",
			ProgramsProvided: []string{"mc"},
			Expression:       `([a-z0-9]{32})-(minio-client)-(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z)`,
		},
	},
	VersionLookupCommand: true,
}

func DiscoverNixStoreExecutables(opts *DiscoverNixStoreOptions) []Executable {
	var result []Executable
	if opts == nil {
		opts = &DefaultDiscoverNixOptions
	}

	// check for nix-store command
	if _, err := exec.LookPath("nix-store"); err != nil {
		log.Debug().Msg("nix-store command not found, skipping nix store discovery")
		return result
	}

	// discover using store paths
	for _, dir := range nixCurrentSystemStorePaths() {
		var hash, pkgName, pkgVersion string
		var programsProvided []string
		var env map[string]string

		for _, pkg := range opts.Packages {
			if hash, pkgName, pkgVersion = nixPathToHashNameVersion(dir, pkg.Expression); hash != "" {
				env = pkg.Env
				programsProvided = pkg.ProgramsProvided
				pkgVersion = util.ToSemanticVersion(pkgVersion)
				break
			}
		}

		if pkgName == "" {
			continue
		}

		for _, executable := range findExecutablesInDirectory(dir + "/bin") {
			// restrict to executables provided by the package if specified
			if len(programsProvided) > 0 && !slices.Contains(programsProvided, executable) {
				continue
			}

			// preprocess env
			for k, v := range env {
				env[k] = strings.ReplaceAll(v, "{{EXECUTABLE_PKG_DIR}}", dir) // placeholder to set e.g. JAVA_HOME to the store dir when running java
			}

			result = append(result, NixStoreCandidate{
				BaseCandidate: BaseCandidate{
					Name:    executable,
					Version: pkgVersion,
					Type:    ExecutionNixStore,
				},
				AbsolutePath:   fmt.Sprintf("%s/bin/%s", dir, executable),
				Package:        pkgName,
				PackageVersion: pkgVersion,
				Env:            env,
			})
		}
	}

	return result
}

// nixCurrentSystemStorePaths returns all paths in the nix store for the current system profile
func nixCurrentSystemStorePaths() []string {
	cmd := exec.Command("nix-store", "--query", "--requisites", "/run/current-system")
	out, err := cmd.Output()
	if err != nil {
		log.Error().Err(err).Msg("failed to execute nix-store command")
		return nil
	}

	paths := strings.Split(string(out), "\n")

	var nonEmptyPaths []string
	for _, path := range paths {
		if path != "" {
			nonEmptyPaths = append(nonEmptyPaths, path)
		}
	}
	return nonEmptyPaths
}

// nixPathToHashNameVersion extracts the hash, name and version from a nix store path
func nixPathToHashNameVersion(path string, expr string) (hash string, name string, version string) {
	re, err := regexp.Compile("^" + expr + "$")
	if err != nil {
		log.Error().Err(err).Str("expression", expr).Msg("failed to compile regex for nix package")
		return "", "", ""
	}

	matches := re.FindStringSubmatch(filepath.Base(path))
	if len(matches) > 1 {
		return matches[1], matches[2], matches[3]
	}

	return "", "", ""
}
