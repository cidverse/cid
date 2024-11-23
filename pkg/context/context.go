package context

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cidverseutils/filesystem"
)

type CIDContext struct {
	// ProjectDir is the root directory of the project
	ProjectDir string

	// WorkDir is the current working directory, must be a subdirectory of ProjectDir or ProjectDir itself
	WorkDir string

	// Config is the CID configuration for the project
	Config *config.CIDConfig

	// Env holds the resolved environment variables for the project
	Env map[string]string
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
	env := api.GetCIDEnvironment(cfg.Env, projectDir)

	// validate
	if !strings.HasPrefix(workDir, projectDir) {
		return nil, ErrWorkDirNotInProjectDir
	}

	return &CIDContext{
		ProjectDir: projectDir,
		WorkDir:    workDir,
		Config:     cfg,
		Env:        env,
	}, nil
}
