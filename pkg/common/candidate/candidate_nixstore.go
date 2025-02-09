package candidate

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
)

// NixStoreCandidate is used for the execution using binaries in the nix store
type NixStoreCandidate struct {
	BaseCandidate
	AbsolutePath   string `yaml:"absolute-path,omitempty"`
	Package        string `yaml:"package,omitempty"`
	PackageVersion string `yaml:"package-version,omitempty"`
}

func (c NixStoreCandidate) GetUri() string {
	return fmt.Sprintf("nix-store:/%s", c.AbsolutePath)
}

func (c NixStoreCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running NixStoreCandidate %s %s with args %v", c.Package, c.PackageVersion, opts.Args)

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

	cmdArgs := append([]string{c.AbsolutePath}, opts.Args...)
	cmd, err := shellcommand.PrepareCommand(strings.Join(cmdArgs, " "), runtime.GOOS, "", true, opts.Env, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
