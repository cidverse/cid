package restapi

import (
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type APIConfig struct {
	BuildID       string
	JobID         string
	ProjectDir    string
	Modules       []*analyzerapi.ProjectModule
	CurrentModule *analyzerapi.ProjectModule
	CurrentAction *config.Action
	Env           map[string]string
	ActionConfig  string
	State         *state.ActionStateContext
	TempDir       string
	ArtifactDir   string
}

type handlerConfig struct {
	buildID             string
	jobID               string
	projectDir          string
	containerProjectDir string
	modules             []*analyzerapi.ProjectModule
	currentModule       *analyzerapi.ProjectModule
	currentAction       *config.Action
	env                 map[string]string
	actionConfig        string
	state               *state.ActionStateContext
	tempDir             string
	artifactDir         string
}

// apiError, see https://www.rfc-editor.org/rfc/rfc7807
type apiError struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}
