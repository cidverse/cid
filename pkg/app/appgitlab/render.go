package appgitlab

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/*
var embedFS embed.FS

type TemplateData struct {
	Version            string                                  `json:"version"`
	Stages             []string                                `json:"stages"`
	Workflows          []appconfig.WorkflowData                `json:"workflows"`
	WorkflowDependency map[string]appconfig.WorkflowDependency `json:"workflow_dependency"`
}

type RenderWorkflowResult struct {
	Plan            plangenerate.Plan
	WorkflowContent string
}

// renderWorkflow renders the workflow template and returns the rendered template and the hash
func renderWorkflow(data []appconfig.WorkflowData, templateFile string, outputFile string) (RenderWorkflowResult, error) {
	content, err := embedFS.ReadFile(path.Join("templates", templateFile))
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to read workflow template %s: %w", templateFile, err)
	}

	var wfStages []string
	wfDependencies := make(map[string]appconfig.WorkflowDependency)
	for _, wf := range data {
		for _, s := range wf.Plan.Stages {
			if !slices.Contains(wfStages, s) {
				wfStages = append(wfStages, s)
			}
		}
		for k, v := range wf.WorkflowDependency {
			wfDependencies[k] = v
		}
	}

	template, err := vcsapp.Render(string(content), TemplateData{
		Version:            constants.Version,
		Stages:             wfStages,
		Workflows:          data,
		WorkflowDependency: wfDependencies,
	})
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

	// TODO: data.Plan
	return RenderWorkflowResult{WorkflowContent: string(template)}, nil
}
