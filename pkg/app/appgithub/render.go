package appgithub

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/gosimple/slug"
)

//go:embed templates/*
var embedFS embed.FS

type WorkflowTemplateData struct {
	Version            string                                  `json:"version"`
	Name               string                                  `json:"name"`
	JobTimeout         int                                     `json:"job_timeout"`
	DefaultBranch      string                                  `json:"default_branch"`
	WorkflowKey        string                                  `json:"workflow_key"`
	WorkflowConfig     appconfig.WorkflowConfig                `json:"workflow_config"`
	Plan               plangenerate.Plan                       `json:"plan"`
	WorkflowDependency map[string]appconfig.WorkflowDependency `json:"workflow_dependency"`
	IgnoreFiles        []string                                `json:"ignore_files"`
}

type RenderWorkflowResult struct {
	Plan            plangenerate.Plan
	WorkflowContent string
}

// renderWorkflow renders the workflow template and returns the rendered template and the hash
func renderWorkflow(cidContext *context.CIDContext, taskContext taskcommon.TaskContext, conf appconfig.Config, wfName string, wfConfig appconfig.WorkflowConfig, environments map[string]appcommon.VCSEnvironment, templateFile string, outputFile string) (RenderWorkflowResult, error) {
	wfConfig = appconfig.PreProcessWorkflowConfig(wfConfig, taskContext.Repository)

	// generate plan
	plan, err := plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
		Modules:         cidContext.Modules,
		Registry:        cidContext.Config.Registry,
		ProjectDir:      taskContext.Directory,
		Env:             cidContext.Env,
		Executables:     cidContext.Executables,
		Environments:    environments,
		PinVersions:     false,
		WorkflowVariant: wfConfig.Type, // workflow variant, e.g. "nightly", "release", "pull-request"
	})
	if err != nil {
		return RenderWorkflowResult{}, err
	}

	// pre-process access section for workflow rendering
	for i := range plan.Steps {
		plan.Steps[i].Access.Network = append(plan.Steps[i].Access.Network, githubNetworkAllowList...)
		plan.Steps[i].Access.Environment = appcommon.RemoveEnvByName(plan.Steps[i].Access.Environment, []string{"GITHUB_TOKEN"})
	}

	// render workflow template
	data := WorkflowTemplateData{
		Version:        constants.Version,
		Name:           wfName,
		JobTimeout:     conf.JobTimeout,
		DefaultBranch:  taskContext.Repository.DefaultBranch,
		WorkflowKey:    slug.Make(wfName),
		WorkflowConfig: wfConfig,
		Plan:           plan,
		WorkflowDependency: map[string]appconfig.WorkflowDependency{
			"actions/checkout": {
				Id:      "actions/checkout",
				Type:    "github-action",
				Version: "v4.2.2",
				Hash:    "11bd71901bbe5b1630ceea73d27597364c9af683",
			},
			"actions/download-artifact": {
				Id:      "actions/download-artifact",
				Type:    "github-action",
				Version: "v4.2.1",
				Hash:    "95815c38cf2ff2164869cbab79da8d1f422bc89e",
			},
			"actions/upload-artifact": {
				Id:      "actions/upload-artifact",
				Type:    "github-action",
				Version: "v4.6.2",
				Hash:    "ea165f8d65b6e75b540449e92b4886f43607fa02",
			},
			"step-security/harden-runner": {
				Id:      "step-security/harden-runner",
				Type:    "github-action",
				Version: "v2.11.1",
				Hash:    "c6295a65d1254861815972266d5933fd6e532bdf",
			},
			"cidverse/ghact-cid-setup": {
				Id:      "cidverse/ghact-cid-setup",
				Type:    "github-action",
				Version: "v0.2.0",
				Hash:    "c6dac0517d28bd8871c195fee9a6bd5a5854d5cb",
			},
		},
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
	content, err := embedFS.ReadFile(path.Join("templates", templateFile))
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to read workflow template %s: %w", templateFile, err)
	}
	template, err := vcsapp.Render(string(content), data)
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to render template %s: %w", templateFile, err)
	}

	// write workflow file
	if outputFile != "" {
		// create workflow file
		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		if err != nil {
			return RenderWorkflowResult{}, fmt.Errorf("failed to create workflow file parent directory: %w", err)
		}
		err = os.WriteFile(outputFile, template, 0644)
		if err != nil {
			return RenderWorkflowResult{}, fmt.Errorf("failed to create workflow file: %w", err)
		}
	}

	return RenderWorkflowResult{Plan: plan, WorkflowContent: string(template)}, nil
}
