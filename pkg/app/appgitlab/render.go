package appgitlab

import (
	"embed"
	"fmt"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"os"
	"path"
	"path/filepath"
	"slices"
)

//go:embed templates/*
var embedFS embed.FS

type TemplateData struct {
	Version                      string                                  `json:"version"`
	ContainerRuntime             string                                  `json:"container_runtime"`
	Stages                       []string                                `json:"stages"`
	Workflows                    []appconfig.WorkflowData                `json:"workflows"`
	WorkflowDependency           map[string]appconfig.WorkflowDependency `json:"workflow_dependency"`
	ReferencedWorkflowDependency map[string]appconfig.WorkflowDependency `json:"-"`
}

func (t *TemplateData) GetDependencyReference(key string) string {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		for _, w := range t.Workflows {
			w.ReferencedWorkflowDependency[key] = dep
		}
		return appconfig.FormatDependencyReference(dep)
	}
	return ""
}

func (t *TemplateData) GetDependency(key string) appconfig.WorkflowDependency {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		for _, w := range t.Workflows {
			w.ReferencedWorkflowDependency[key] = dep
		}
		return dep
	}
	return appconfig.WorkflowDependency{}
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

	var containerRuntime string
	var wfStages []string
	wfDependencies := make(map[string]appconfig.WorkflowDependency)
	for _, wf := range data {
		containerRuntime = wf.ContainerRuntime
		for _, s := range wf.Plan.Stages {
			if !slices.Contains(wfStages, s) {
				wfStages = append(wfStages, s)
			}
		}
		for k, v := range wf.WorkflowDependency {
			wfDependencies[k] = v
		}
	}
	wfStages = getOrderedStages(wfStages) // TODO: this is not ideal, but returns the correct order for now

	templateData := &TemplateData{
		Version:                      constants.Version,
		ContainerRuntime:             containerRuntime,
		Stages:                       wfStages,
		Workflows:                    data,
		WorkflowDependency:           wfDependencies,
		ReferencedWorkflowDependency: make(map[string]appconfig.WorkflowDependency),
	}
	template, err := vcsapp.Render(string(content), templateData)
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to render template %s: %w", templateFile, err)
	}
	templateString := string(template)

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

	return RenderWorkflowResult{WorkflowContent: templateString}, nil
}
