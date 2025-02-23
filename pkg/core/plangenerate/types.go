package plangenerate

import (
	"errors"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

var (
	ErrNoSuitableWorkflowFound = errors.New("no suitable workflow found")
)

type Plan struct {
	Steps []Step `json:"steps"`
}

type Stage struct {
	Name  string   `json:"name"`
	Steps []Step   `json:"steps,omitempty"`
	Needs []string `json:"needs,omitempty"`
}

type Step struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Stage    string              `json:"stage"`
	Scope    catalog.ActionScope `json:"scope"`
	Action   string              `json:"action"`
	Module   string              `json:"module,omitempty"`
	RunAfter []string            `json:"run-after,omitempty"`
	Order    int                 `json:"order"`
}

type Action struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

type PlanContext struct {
	ProjectDir  string
	Environment map[string]string
	Registry    catalog.Config
	Modules     []*analyzerapi.ProjectModule
}
