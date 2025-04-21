package executable

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
)

// NixShellCandidate is used for the execution using nix-shell
type NixShellCandidate struct {
	BaseCandidate
	Package            string            `yaml:"package,omitempty"`
	PackageVersion     string            `yaml:"package-version,omitempty"`
	AdditionalPackages []string          `yaml:"additional-packages,omitempty"`
	Channel            string            `yaml:"channel,omitempty"`
	Env                map[string]string `json:"env,omitempty"`
}

func (c NixShellCandidate) GetUri() string {
	return fmt.Sprintf("nix-shell:/%s#%s@%s", c.Channel, c.Package, c.Version)
}

func (c NixShellCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running NixShellCandidate %s %s with args %v", c.Package, c.PackageVersion, opts.Args)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	var stdoutWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), &stdoutBuffer)
	var stderrWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stderr, &sync.Mutex{}, nil), &stderrBuffer)
	if opts.HideStdOut {
		stdoutWriter = &stdoutBuffer
	}
	if opts.HideStdErr {
		stderrWriter = &stderrBuffer
	}

	var nixShellArgs = []string{"nix-shell"}
	if c.Channel == "unstable" {
		nixShellArgs = append(nixShellArgs, "-I", "nixpkgs=nixos-unstable")
	}
	nixShellArgs = append(nixShellArgs, "-p", c.Package)
	if len(c.AdditionalPackages) > 0 {
		nixShellArgs = append(nixShellArgs, "-p", strings.Join(c.AdditionalPackages, " "))
	}
	nixShellArgs = append(nixShellArgs, "--run", fmt.Sprintf("%q", strings.Join(opts.Args, " ")))

	env := util.MergeMaps(c.Env, opts.Env)
	env = util.ResolveEnvMap(env)
	env["NIX_PATH"] = os.Getenv("NIX_PATH")
	cmd, err := shellcommand.PrepareCommand(strings.Join(nixShellArgs, " "), runtime.GOOS, "", false, env, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
