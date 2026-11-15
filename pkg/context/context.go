package context

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type CIDContext struct {
	ProjectDir  string                       // ProjectDir is the root directory of the project
	WorkDir     string                       // WorkDir is the current working directory, must be a subdirectory of ProjectDir or ProjectDir itself
	Config      *config.CIDConfig            // Config is the CID configuration for the project
	Env         map[string]string            // Env holds the resolved environment variables for the project
	Modules     []*analyzerapi.ProjectModule // Modules is a list of all discovered modules in the project
	Executables []executable.Executable      // Executables is a list of all executable candidates usable for command execution
}

var (
	ErrProjectDirNotFound     = fmt.Errorf("could not determine project directory")
	ErrWorkDirNotFound        = fmt.Errorf("could not determine current working directory")
	ErrWorkDirNotInProjectDir = fmt.Errorf("workDir must be the projectDir or a subdirectory of projectDir")
)

// NewAppContextFromDir creates the context for running cid in projectDir; it sets config.Current and uses the process environment.
func NewAppContextFromDir(projectDir string, workDir string) (*CIDContext, error) {
	return newAppContext(projectDir, workDir, false)
}

// NewIsolatedAppContextFromDir creates the context for rendering workflows of the project in projectDir (cid app-run, cid app-server).
// It is safe for concurrent use: config.Current is not modified, and neither the process environment nor a .env file in the
// working directory affect the result.
func NewIsolatedAppContextFromDir(projectDir string) (*CIDContext, error) {
	return newAppContext(projectDir, projectDir, true)
}

func newAppContext(projectDir string, workDir string, isolated bool) (*CIDContext, error) {
	slog.With("dir", projectDir).With("isolated", isolated).Debug("initializing cid context")
	var cfg *config.CIDConfig
	if isolated {
		cfg = config.Load(projectDir)
	} else {
		cfg = config.LoadConfig(projectDir)
	}

	// env
	env, err := api.GetCIDEnvironment(api.CIDEnvironmentOptions{
		ProjectDirectory: projectDir,
		ConfigEnv:        cfg.Env,
		Conventions:      cfg.Conventions,
		Isolated:         isolated,
	})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to prepare cid environment"), err)
	}

	// modules
	modules := analyzer.ScanDirectory(projectDir)

	// get candidates
	executables, err := command.CandidatesFromConfig(*cfg)
	if err != nil {
		return nil, err
	}

	return &CIDContext{
		ProjectDir:  projectDir,
		WorkDir:     workDir,
		Config:      cfg,
		Env:         env,
		Modules:     modules,
		Executables: executables,
	}, nil
}

func NewAppContext() (*CIDContext, error) {
	projectDir, err := filesystem.GetProjectDirectory()
	if err != nil {
		return nil, errors.Join(ErrProjectDirNotFound, err)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Join(ErrWorkDirNotFound, err)
	}

	// validate
	if !strings.HasPrefix(workDir, projectDir) {
		return nil, ErrWorkDirNotInProjectDir
	}

	return NewAppContextFromDir(projectDir, workDir)
}
