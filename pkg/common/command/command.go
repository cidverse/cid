package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/go-ptr"
)

var (
	ErrNoCandidatesProvided = fmt.Errorf("candidates are a required field")
	ErrNoCommandProvided    = fmt.Errorf("command is a required field")
)

type Opts struct {
	Candidates             []executable.Candidate
	CandidateTypes         []executable.CandidateType
	Command                string
	Env                    map[string]string
	ProjectDir             string
	WorkDir                string
	TempDir                string
	CaptureOutput          bool
	Ports                  []int
	UserProvidedConstraint string
	Constraints            map[string]string
	Stdin                  io.Reader
}

// Execute gets called from actions or the api to execute commands
func Execute(opts Opts) (stdout string, stderr string, cand executable.Candidate, err error) {
	// validate
	if len(opts.Candidates) == 0 {
		return "", "", cand, ErrNoCandidatesProvided
	}
	if len(opts.Command) == 0 {
		return "", "", cand, ErrNoCommandProvided
	}

	// identify command
	args := strings.SplitN(opts.Command, " ", 2)
	cmdBinary := args[0]

	// constraint from config
	versionConstraint := executable.AnyVersionConstraint
	if value, ok := opts.Constraints[cmdBinary]; ok {
		versionConstraint = value
	}
	// user provided constraint
	if len(opts.UserProvidedConstraint) > 0 {
		versionConstraint = opts.UserProvidedConstraint
	}

	// select candidate
	c := executable.SelectCandidate(opts.Candidates, executable.CandidateFilter{
		Types:             opts.CandidateTypes,
		Executable:        cmdBinary,
		VersionPreference: executable.PreferHighest,
		VersionConstraint: versionConstraint,
	})
	if c == nil {
		return "", "", nil, fmt.Errorf("no candidate found for %s fulfilling constraint %s", cmdBinary, versionConstraint)
	}
	cand = ptr.Value(c)

	// run command
	stdout, stderr, err = cand.Run(executable.RunParameters{
		Executable:    cmdBinary,
		Args:          args,
		Env:           opts.Env,
		RootDir:       opts.ProjectDir,
		WorkDir:       opts.WorkDir,
		TempDir:       opts.TempDir,
		CaptureOutput: opts.CaptureOutput,
	})
	if err != nil {
		return stdout, stderr, cand, err
	}

	return stdout, stderr, cand, nil
}
