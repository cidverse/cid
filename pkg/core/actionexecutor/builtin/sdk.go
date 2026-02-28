package builtin

import (
	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type ActionSDK struct {
	BuildID              string
	JobID                string
	ProjectDir           string
	Modules              []*analyzerapi.ProjectModule
	Step                 plangenerate.Step
	CurrentModule        *analyzerapi.ProjectModule
	CurrentAction        *catalog.Action
	NCI                  nci.Spec
	Env                  map[string]string
	ActionEnv            map[string]string
	ActionConfig         string
	State                *state.ActionStateContext
	TempDir              string
	ArtifactDir          string
	ExecutableCandidates []executable.Executable
}
