package executable

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
)

// NixShellCandidate is used for the execution using nix-shell
type NixShellCandidate struct {
	BaseCandidate
	Package            string   `yaml:"package,omitempty"`
	PackageVersion     string   `yaml:"package-version,omitempty"`
	AdditionalPackages []string `yaml:"additional-packages,omitempty"`
	Channel            string   `yaml:"channel,omitempty"`
	//nix-shell with channel: nix-shell -I nixpkgs=channel:nixos-unstable -p hugo
}

func (c NixShellCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running NixShellCandidate %s %s with args %v", c.Package, c.PackageVersion, opts.Args)

	var stdoutWriter io.Writer
	var stderrWriter io.Writer
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if opts.CaptureOutput {
		stdoutWriter = redact.NewProtectedWriter(nil, &stdoutBuffer, &sync.Mutex{}, nil)
		stderrWriter = redact.NewProtectedWriter(nil, &stderrBuffer, &sync.Mutex{}, nil)
	} else {
		stdoutWriter = redact.NewProtectedWriter(os.Stdout, nil, &sync.Mutex{}, nil)
		stderrWriter = redact.NewProtectedWriter(os.Stderr, nil, &sync.Mutex{}, nil)
	}

	cmd, err := shellcommand.PrepareCommand("", runtime.GOOS, "bash", true, nil, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
