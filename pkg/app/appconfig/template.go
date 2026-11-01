package appconfig

import (
	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/gosimple/slug"
)

type WorkflowData struct {
	Version                      string                           `json:"version"`
	Name                         string                           `json:"name"`
	NameSlug                     string                           `json:"name_slug"`
	JobTimeout                   int                              `json:"job_timeout"`
	DefaultBranch                string                           `json:"default_branch"`
	ContainerRuntime             string                           `json:"container_runtime"`
	WorkflowKey                  string                           `json:"workflow_key"`
	WorkflowConfig               WorkflowConfig                   `json:"workflow_config"`
	NetworkGlobal                plangenerate.StepAccessNetwork   `json:"network_global,omitempty"` // NetworkGlobal is the network allow list shared by all steps of the workflow
	Plan                         plangenerate.Plan                `json:"plan"`
	WorkflowDependency           map[string]dependency.Dependency `json:"-"`
	ReferencedWorkflowDependency map[string]dependency.Dependency `json:"workflow_dependency"`
	IgnoreFiles                  []string                         `json:"ignore_files"`
}

func (t *WorkflowData) GetDependencyReference(key string) string {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		return dep.AsDependencyReference()
	}
	return ""
}

func (t *WorkflowData) GetDependency(key string) dependency.Dependency {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		return dep
	}
	return dependency.Dependency{}
}

// AllowedEndpoints returns all network endpoints a step may use, including the workflow-wide allow list
func (t *WorkflowData) AllowedEndpoints(step plangenerate.Step) plangenerate.StepAccessNetwork {
	return plangenerate.MergeNetworkHosts(step.Access.Network, t.NetworkGlobal)
}

type RenderWorkflowResult struct {
	Plan            plangenerate.Plan
	WorkflowContent string
}

// GenerateWorkflowData generates the workflow template data
func GenerateWorkflowData(cidContext *context.CIDContext, taskContext taskcommon.TaskContext, conf Config, wfName string, wfConfig WorkflowConfig, vars []api.CIVariable, environments map[string]appcommon.VCSEnvironment, wfDependencies map[string]dependency.Dependency, networkAllowGlobal []actionsdk.ActionAccessNetwork) (WorkflowData, error) {
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

	// the global network allow list is stored once per workflow, steps only keep their own endpoints
	networkGlobal := plangenerate.MergeNetworkHosts(plangenerate.NetworkHosts(networkAllowGlobal))
	for i := range plan.Steps {
		plan.Steps[i].Access.Network = plangenerate.RemoveNetworkHosts(plan.Steps[i].Access.Network, networkGlobal)
	}

	// render workflow template
	data := WorkflowData{
		Version:                      constants.Version,
		Name:                         wfName,
		NameSlug:                     slug.Make(wfName),
		JobTimeout:                   conf.JobTimeout,
		DefaultBranch:                taskContext.Repository.DefaultBranch,
		ContainerRuntime:             conf.ContainerRuntime,
		WorkflowKey:                  slug.Make(wfName),
		WorkflowConfig:               wfConfig,
		Plan:                         plan,
		NetworkGlobal:                networkGlobal,
		WorkflowDependency:           wfDependencies,
		ReferencedWorkflowDependency: make(map[string]dependency.Dependency),
		IgnoreFiles: []string{
			"README.md",
			"LICENSE",
			".gitignore",
			".gitattributes",
			".editorconfig",
			"renovate.json",
			"CODEOWNERS",
			"SECURITY.md",
		},
	}

	return data, nil
}
