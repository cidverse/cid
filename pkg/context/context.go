package context

import (
	"errors"
	"fmt"
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

func NewAppContext() (*CIDContext, error) {
	projectDir, err := filesystem.GetProjectDirectory()
	if err != nil {
		return nil, errors.Join(ErrProjectDirNotFound, err)
	}
	workDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Join(ErrWorkDirNotFound, err)
	}
	cfg := config.LoadConfig(projectDir)

	// env
	env, err := api.GetCIDEnvironment(cfg.Env, projectDir)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to prepare cid environment"), err)
	}

	// validate
	if !strings.HasPrefix(workDir, projectDir) {
		return nil, ErrWorkDirNotInProjectDir
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
