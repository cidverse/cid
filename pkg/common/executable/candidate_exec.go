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

// ExecCandidate is used for the execution using locally installed binaries
type ExecCandidate struct {
	BaseCandidate
	AbsolutePath string            `json:"absolute-path,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
}

func (c ExecCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running ExecCandidate %s with args %v", c.AbsolutePath, opts.Args)

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

	// replace executable with absolute path
	if opts.Args[0] == opts.Executable {
		opts.Args[0] = c.AbsolutePath
	}

	env := util.MergeMaps(c.Env, opts.Env)
	env = util.ResolveEnvMap(env)
	cmd, err := shellcommand.PrepareCommand(strings.Join(opts.Args, " "), runtime.GOOS, "bash", false, env, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
