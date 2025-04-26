package restapi

import (
	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type APIConfig struct {
	BuildID              string
	JobID                string
	ProjectDir           string
	Modules              []*analyzerapi.ProjectModule
	Step                 plangenerate.Step
	CurrentModule        *analyzerapi.ProjectModule
	CurrentAction        *catalog.Action
	Env                  map[string]string
	ActionEnv            map[string]string
	ActionConfig         string
	State                *state.ActionStateContext
	TempDir              string
	ArtifactDir          string
	ExecutableCandidates []executable.Executable
}

// apiError, see https://www.rfc-editor.org/rfc/rfc7807
type apiError struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}
