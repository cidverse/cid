package appgitlab

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/*
var embedFS embed.FS

type TemplateData struct {
	Version   string                   `json:"version"`
	Stages    []string                 `json:"stages"`
	Workflows []appconfig.WorkflowData `json:"workflows"`
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
	template, err := vcsapp.Render(string(content), TemplateData{
		Version:   constants.Version,
		Stages:    data[0].Plan.Stages, // stages should be the same for all workflows
		Workflows: data,
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
