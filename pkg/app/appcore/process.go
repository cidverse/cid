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

func ProcessRepository(platform api.Platform, repo api.Repository, dryRun bool) (apptask.WorkflowTaskResult, error) {
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
		return apptask.WorkflowTaskResult{}, appgithub.GitHubWorkflowTask(taskContext)
	} else if platform.Slug() == "gitlab" {
		return appgitlab.GitLabWorkflowTask(taskContext, dryRun)
	} else {
		return apptask.WorkflowTaskResult{}, fmt.Errorf("platform %s not supported", platform.Slug())
	}
}
