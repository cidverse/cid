package plangenerate

import (
	"errors"
	"fmt"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"strconv"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
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
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	Slug               string               `json:"slug"`
	Stage              string               `json:"stage"`
	Scope              catalog.ActionScope  `json:"scope"`
	Action             string               `json:"action"`
	Module             string               `json:"module,omitempty"`
	RunAfter           []string             `json:"run-after,omitempty"`              // List of steps that need to be completed before this step starts (by slug)
	RunAfterByName     []string             `json:"run-after-by-name,omitempty"`      // List of steps that need to be completed before this step starts (by name)
	UsesOutputOf       []string             `json:"uses-output-of,omitempty"`         // List of steps whose outputs need to be downloaded (by slug)
	UsesOutputOfByName []string             `json:"uses-output-of-by-name,omitempty"` // List of steps whose outputs need to be downloaded (by name)
	Environment        string               `json:"environment,omitempty"`
	Access             catalog.ActionAccess `json:"access,omitempty"`
	Inputs             catalog.ActionInput  `json:"inputs,omitempty"`
	Outputs            catalog.ActionOutput `json:"outputs,omitempty"`
	Order              int                  `json:"order"` // Topological order
	Config             interface{}          `json:"config,omitempty"`
}

func (s *Step) HasOutputWithTypeAndFormat(artifactType string, artifactFormat string) bool {
	for _, artifact := range s.Outputs.Artifacts {
		if artifact.Type == artifactType && artifact.Format == artifactFormat {
			return true
		}
	}

	return false
}

func buildStep(catalogAction catalog.Action, action catalog.WorkflowAction, id int, name string, moduleRef *analyzerapi.ProjectModule, environment string, executableConstraints []catalog.ActionAccessExecutable) Step {
	moduleName := ""
	if moduleRef != nil {
		moduleName = moduleRef.ID
		name = fmt.Sprintf("%s [%s]", name, moduleRef.Name)
	}
	if environment != "" {
		name = fmt.Sprintf("%s (%s)", name, environment)
	}

	return Step{
		ID:          strconv.Itoa(id),
		Name:        name,
		Slug:        slug.Make(name),
		Stage:       action.Stage,
		Scope:       catalogAction.Metadata.Scope,
		Module:      moduleName,
		Action:      catalogAction.URI,
		RunAfter:    []string{},
		Environment: environment,
		Access: catalog.ActionAccess{
			Environment: catalogAction.Metadata.Access.Environment,
			Executables: executableConstraints,
			Network:     catalogAction.Metadata.Access.Network,
		},
		Inputs:  catalogAction.Metadata.Input,
		Outputs: catalogAction.Metadata.Output,
		Order:   1,
		Config:  action.Config,
	}
}

type StepMetadata struct {
	ExecutableConstraints map[string]string `json:"executable-constraints,omitempty"`
}

type PlanContext struct {
	ProjectDir      string
	Environment     map[string]string
	Stages          []string
	VCSVariables    []api.CIVariable
	VCSEnvironments map[string]appcommon.VCSEnvironment
	Registry        catalog.Config
	Modules         []*analyzerapi.ProjectModule
}
