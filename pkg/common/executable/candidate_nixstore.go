package executable

import (
	"bytes"
	"errors"
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

// NixStoreCandidate is used for the execution using binaries in the nix store
type NixStoreCandidate struct {
	BaseCandidate
	AbsolutePath   string            `json:"absolute-path,omitempty"`
	Package        string            `json:"package,omitempty"`
	PackageVersion string            `json:"package-version,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
}

func (c NixStoreCandidate) GetUri() string {
	return fmt.Sprintf("nix-store:/%s", c.AbsolutePath)
}

func (c NixStoreCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running NixStoreCandidate %s %s with args %v", c.Package, c.PackageVersion, opts.Args)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	var stdoutWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), &stdoutBuffer)
	var stderrWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stderr, &sync.Mutex{}, nil), &stderrBuffer)
	if opts.HideStdOut {
		stdoutWriter = &stdoutBuffer
	}
	if opts.HideStdErr {
		stderrWriter = &stderrBuffer
	}

	// replace executable with absolute path
	if opts.Args[0] == opts.Executable {
		if _, err := os.Stat(c.AbsolutePath); err != nil {
			if os.IsNotExist(err) {
				return "", "", errors.Join(ErrCheckingForExecutable, fmt.Errorf("path: %q", c.AbsolutePath))
			}
			return "", "", errors.Join(ErrExecutableNotFound, fmt.Errorf("path: %q", c.AbsolutePath))
		}
		opts.Args[0] = c.AbsolutePath
	}

	env := util.MergeMaps(c.Env, opts.Env)
	env = util.ResolveEnvMap(env)
	cmd, err := shellcommand.PrepareCommand(strings.Join(opts.Args, " "), runtime.GOOS, "", false, env, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
