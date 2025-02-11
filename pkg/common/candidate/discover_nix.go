package candidate

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

const nixStorePath = "/nix/store"

type NixPackage struct {
	Name       string
	Expression string
	Env        map[string]string
}

type DiscoverNixOptions struct {
	Packages             []NixPackage
	VersionLookupCommand bool
}

var DefaultDiscoverNixOptions = DiscoverNixOptions{
	Packages: []NixPackage{
		{
			Name:       "openjdk",
			Expression: `([a-z0-9]{32})-(openjdk)-(\d+\.\d+\.\d+.+)`,
		},
		{
			Name:       "go",
			Expression: `([a-z0-9]{32})-(go)-(\d+\.\d+\.\d+)`,
			Env: map[string]string{
				"GOPATH":     "$HOME/go",
				"GOMODCACHE": "$HOME/go/pkg/mod",
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
			Name:       "openshift",
			Expression: `([a-z0-9]{32})-(openshift)-(\d+\.\d+\.\d+)`,
		},
		{
			Name:       "ansible",
			Expression: `([a-z0-9]{32})-(ansible)-(\d+\.\d+\.\d+)`,
		},
	},
	VersionLookupCommand: true,
}

func DiscoverNixStoreCandidates(opts *DiscoverNixOptions) []Candidate {
	var result []Candidate
	if opts == nil {
		opts = &DefaultDiscoverNixOptions
	}

	// discover using store paths
	for _, dir := range nixCurrentSystemStorePaths() {
		var hash, pkgName, pkgVersion string
		var env map[string]string

		for _, pkg := range opts.Packages {
			if hash, pkgName, pkgVersion = nixPathToHashNameVersion(dir, pkg.Expression); hash != "" {
				env = pkg.Env
				break
			}
		}

		if pkgName == "" {
			continue
		}

		for _, executable := range findExecutablesInDirectory(dir + "/bin") {
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
