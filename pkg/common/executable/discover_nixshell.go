package executable

import (
	"os/exec"

	"github.com/cidverse/cid/pkg/util"
	"github.com/rs/zerolog/log"
)

type NixShellPackage struct {
	Name             string
	ProgramsProvided []string
	Env              map[string]string
}

type DiscoverNixShellOptions struct {
	Packages             []NixShellPackage
	VersionLookupCommand bool
}

var DefaultDiscoverNixShellOptions = DiscoverNixShellOptions{
	Packages: []NixShellPackage{
		{
			Name:             "openjdk",
			ProgramsProvided: []string{"java", "javac", "javadoc", "javap"},
			Env: map[string]string{
				"JAVA_HOME": "{{EXECUTABLE_PKG_DIR}}/lib/openjdk",
			},
		},
		{
			Name:             "go",
			ProgramsProvided: []string{"go"},
			Env: map[string]string{
				"GOPATH":     "$HOME/go",
				"GOMODCACHE": "$HOME/go/pkg/mod",
				"GOCACHE":    "$HOME/.cache/go-build",
			},
		},
		{
			Name:             "zig",
			ProgramsProvided: []string{"zig"},
		},
		{
			Name:             "kubernetes-helm",
			ProgramsProvided: []string{"helm"},
		},
		{
			Name:             "helmfile",
			ProgramsProvided: []string{"helmfile"},
		},
		{
			Name:             "kubectl",
			ProgramsProvided: []string{"kubectl"},
		},
		{
			Name:             "openshift",
			ProgramsProvided: []string{"oc"},
		},
		{
			Name:             "ansible",
			ProgramsProvided: []string{"ansible"},
		},
		{
			Name:             "minio-client",
			ProgramsProvided: []string{"mc"},
		},
	},
	VersionLookupCommand: true,
}

func DiscoverNixShellExecutables(opts *DiscoverNixShellOptions) []Executable {
	var result []Executable
	if opts == nil {
		opts = &DefaultDiscoverNixShellOptions
	}

	// check for nix-store command
	if _, err := exec.LookPath("nix-shell"); err != nil {
		log.Debug().Msg("nix-shell command not found, skipping nix shell discovery")
		return result
	}

	for _, p := range opts.Packages {
		channel := "nixpkgs"

		// eval to get pkg version
		cmd := exec.Command("nix", "eval", "--raw", channel+"#"+p.Name+".version")
		version, err := cmd.Output()
		if err != nil {
			log.Debug().Err(err).Str("package", p.Name).Msg("failed to get package version")
			continue
		}

		// add to result
		for _, pp := range p.ProgramsProvided {
			result = append(result, NixShellCandidate{
				BaseCandidate: BaseCandidate{
					Name:    pp,
					Version: util.ToSemanticVersion(string(version)),
					Type:    ExecutionNixShell,
				},
				Package: p.Name,
				Channel: channel,
				Env:     p.Env,
			})
		}
	}

	return result
}
