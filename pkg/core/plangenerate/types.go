package plangenerate

import (
	"errors"

	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

var (
	ErrNoSuitableWorkflowFound = errors.New("no suitable workflow found")
)

type Plan struct {
	Name              string                  `json:"name"`
	Stages            []string                `json:"stages"`
	Steps             []Step                  `json:"steps"`
	PinnedExecutables []executable.Executable `json:"pinned-executables,omitempty"`
}

type Stage struct {
	Name  string   `json:"name"`
	Steps []Step   `json:"steps,omitempty"`
	Needs []string `json:"needs,omitempty"`
}

type Step struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Stage    string               `json:"stage"`
	Scope    catalog.ActionScope  `json:"scope"`
	Action   string               `json:"action"`
	Module   string               `json:"module,omitempty"`
	RunAfter []string             `json:"run-after,omitempty"`
	Access   catalog.ActionAccess `json:"access,omitempty"`
	Order    int                  `json:"order"` // Topological order
	Config   interface{}          `json:"config,omitempty"`
}

type StepMetadata struct {
	ExecutableConstraints map[string]string `json:"executable-constraints,omitempty"`
}

type PlanContext struct {
	ProjectDir  string
	Environment map[string]string
	Registry    catalog.Config
	Modules     []*analyzerapi.ProjectModule
}
