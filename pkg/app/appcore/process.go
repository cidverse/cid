package appcore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appgithub"
	"github.com/cidverse/cid/pkg/app/appgitlab"
	"github.com/cidverse/cid/pkg/app/apptask"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/gosimple/slug"
)

// ProcessOptions controls how a repository is processed.
type ProcessOptions struct {
	// DryRun renders the workflows without committing or opening a merge request.
	DryRun bool
	// SelfContained renders GitLab pipelines that need no committed .cid/ files (used by cid app-server). Ignored for GitHub.
	SelfContained bool
}

func ProcessRepository(platform api.Platform, repo api.Repository, opts ProcessOptions) (apptask.WorkflowTaskResult, error) {
	// create temp directory
	tempDir, err := os.MkdirTemp("", "cid-app-*")
	if err != nil {
		return apptask.WorkflowTaskResult{}, fmt.Errorf("failed to prepare temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// run platform-specific task
	taskContext := taskcommon.TaskContext{
		Directory:  filepath.Join(tempDir, slug.Make(repo.Name)),
		Platform:   platform,
		Repository: repo,
	}
	err = os.MkdirAll(taskContext.Directory, os.ModePerm)
	if err != nil {
		return apptask.WorkflowTaskResult{}, fmt.Errorf("failed to create directory: %w", err)
	}
	if platform.Slug() == "github" {
		return appgithub.GitHubWorkflowTask(taskContext, opts.DryRun)
	} else if platform.Slug() == "gitlab" {
		return appgitlab.GitLabWorkflowTask(taskContext, opts.DryRun, opts.SelfContained)
	} else {
		return apptask.WorkflowTaskResult{}, fmt.Errorf("platform %s not supported", platform.Slug())
	}
}
