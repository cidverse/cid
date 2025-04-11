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
)

//go:embed templates/*
var embedFS embed.FS

type WorkflowTemplateData struct {
	Version        string                   `json:"version"`
	Name           string                   `json:"name"`
	JobTimeout     int                      `json:"job_timeout"`
	DefaultBranch  string                   `json:"default_branch"`
	WorkflowConfig appconfig.WorkflowConfig `json:"workflow_config"`
	Plan           plangenerate.Plan        `json:"plan"`
	IgnoreFiles    []string                 `json:"ignore_files"`
}

// renderWorkflow renders the workflow template and returns the rendered template and the hash
func renderWorkflow(cidContext *context.CIDContext, taskContext taskcommon.TaskContext, conf appconfig.Config, wfName string, wfConfig appconfig.WorkflowConfig, environments map[string]appcommon.VCSEnvironment, templateFile string, outputFile string) (string, error) {
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
		return "", err
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
		WorkflowConfig: wfConfig,
		Plan:           plan,
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
		return "", fmt.Errorf("failed to read workflow template %s: %w", templateFile, err)
	}
	template, err := vcsapp.Render(string(content), data)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateFile, err)
	}

	// write workflow file
	if outputFile != "" {
		// create workflow file
		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create workflow file parent directory: %w", err)
		}
		err = os.WriteFile(outputFile, template, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to create workflow file: %w", err)
		}
	}

	return string(template), nil
}
