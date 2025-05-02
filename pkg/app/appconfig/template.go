package appconfig

import (
	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/gosimple/slug"
)

type WorkflowData struct {
	Version            string                        `json:"version"`
	Name               string                        `json:"name"`
	NameSlug           string                        `json:"name_slug"`
	JobTimeout         int                           `json:"job_timeout"`
	DefaultBranch      string                        `json:"default_branch"`
	WorkflowKey        string                        `json:"workflow_key"`
	WorkflowConfig     WorkflowConfig                `json:"workflow_config"`
	Plan               plangenerate.Plan             `json:"plan"`
	WorkflowDependency map[string]WorkflowDependency `json:"workflow_dependency"`
	IgnoreFiles        []string                      `json:"ignore_files"`
}

type TemplateData struct {
	Version   string         `json:"version"`
	Stages    []string       `json:"stages"`
	Workflows []WorkflowData `json:"workflows"`
}

type RenderWorkflowResult struct {
	Plan            plangenerate.Plan
	WorkflowContent string
}

// GenerateWorkflowData generates the workflow template data
func GenerateWorkflowData(cidContext *context.CIDContext, taskContext taskcommon.TaskContext, conf Config, wfName string, wfConfig WorkflowConfig, vars []api.CIVariable, environments map[string]appcommon.VCSEnvironment, wfDependencies map[string]WorkflowDependency, networkAllowGlobal []catalog.ActionAccessNetwork) (WorkflowData, error) {
	wfConfig = PreProcessWorkflowConfig(wfConfig, taskContext.Repository)

	// generate plan
	plan, err := plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
		Modules:      cidContext.Modules,
		Registry:     cidContext.Config.Registry,
		ProjectDir:   taskContext.Directory,
		Env:          cidContext.Env,
		Executables:  cidContext.Executables,
		Variables:    vars,
		Environments: environments,
		PinVersions:  false,
		WorkflowType: wfConfig.Type,
	})
	if err != nil {
		return WorkflowData{}, err
	}

	// pre-process access section for workflow rendering
	for i := range plan.Steps {
		plan.Steps[i].Access.Network = append(plan.Steps[i].Access.Network, networkAllowGlobal...)
		plan.Steps[i].Access.Environment = appcommon.RemoveEnvByName(plan.Steps[i].Access.Environment, []string{"GITHUB_TOKEN"})
	}

	// render workflow template
	data := WorkflowData{
		Version:            constants.Version,
		Name:               wfName,
		NameSlug:           slug.Make(wfName),
		JobTimeout:         conf.JobTimeout,
		DefaultBranch:      taskContext.Repository.DefaultBranch,
		WorkflowKey:        slug.Make(wfName),
		WorkflowConfig:     wfConfig,
		Plan:               plan,
		WorkflowDependency: wfDependencies,
		IgnoreFiles: []string{
			"README.md",
			"LICENSE",
			".gitignore",
			".gitattributes",
			".editorconfig",
			"renovate.json",
			"CODEOWNERS",
		},
	}

	return data, nil
}
