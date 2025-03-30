package appgithub

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appmergerequest"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/gosimple/slug"
	"gopkg.in/yaml.v3"
)

var githubNetworkAllowList = []catalog.ActionAccessNetwork{
	// GitHub Platform
	{Host: "github.com:443"},
	{Host: "api.github.com:443"},
	{Host: "codeload.github.com:443"},
	{Host: "uploads.github.com:443"},
	{Host: "objects.githubusercontent.com:443"},
	{Host: "raw.githubusercontent.com:443"},
	// GitHub Container Registry
	{Host: "ghcr.io:443"},
	{Host: "pkg-containers.githubusercontent.com:443"},
}

func GitHubWorkflowTask(taskContext taskcommon.TaskContext) error {
	helper := simpletask.New(taskContext)
	var generatedContent string

	// read config
	conf := appconfig.Config{
		Version:      constants.Version,
		JobTimeout:   10,
		EgressPolicy: "block",
		Workflows: map[string]appconfig.WorkflowConfig{
			"Main": {
				Type:                "main",
				TriggerManual:       true,
				TriggerPush:         true,
				TriggerPushBranches: []string{taskContext.Repository.DefaultBranch},
			},
			"Release": {
				Type:            "release",
				TriggerManual:   true,
				TriggerPush:     true,
				TriggerPushTags: []string{"v*.*.*"},
			},
			"Pull Request": {
				Type:               "pull-request",
				TriggerManual:      true,
				TriggerPullRequest: true,
			},
			"Nightly": {
				Type:                "nightly",
				TriggerManual:       true,
				TriggerSchedule:     true,
				TriggerScheduleCron: "@daily",
			},
		},
	}
	content, err := taskContext.Platform.FileContent(taskContext.Repository, taskContext.Repository.DefaultBranch, appcommon.ConfigFileName)
	if err == nil {
		err = yaml.Unmarshal([]byte(content), &conf)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// clone repository
	slog.With("dir", taskContext.Directory).With("clone-uri", taskContext.Repository.CloneSSH).Debug("cloning repository")
	err = helper.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// create and checkout new branch
	branch := fmt.Sprintf("%s-%s", appcommon.BranchName, conf.Version)
	err = helper.CreateBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// app context
	cid, err := context.NewAppContextFromDir(taskContext.Directory, taskContext.Directory)
	if err != nil {
		return err
	}

	// workflows
	if conf.Workflows != nil {
		for wfKey, wfConfig := range conf.Workflows {
			data, err := renderWorkflow(cid, taskContext, conf, wfKey, wfConfig, "wf-main.gohtml", filepath.Join(taskContext.Directory, fmt.Sprintf(".github/workflows/cid-%s.yml", slug.Make(wfKey))))
			if err != nil {
				return fmt.Errorf("failed to render workflow: %w", err)
			}
			generatedContent += data
		}
	}

	// description
	/*
		cl, err := changelog.GetChangelog()
		if err != nil {
			return fmt.Errorf("failed to get changelog: %w", err)
		}*/
	title, description, err := appmergerequest.TitleAndDescription("0.0.0", conf.Version, []string{"workflow", "githubactions"}, nil) // TODO: previous version lookup
	if err != nil {
		return fmt.Errorf("failed to get merge request description: %w", err)
	}

	// hash workflow
	workflowHash, err := hash.SHA256Hash(bytes.NewReader([]byte(generatedContent)))
	if err != nil {
		return fmt.Errorf("failed to hash workflow contents: %w", err)
	}

	// commit push and create or update merge request
	err = helper.CommitPushAndMergeRequest(title, title, description, fmt.Sprintf("%s-%s", appcommon.MergeRequestId, workflowHash))
	if err != nil {
		return fmt.Errorf("failed to commit push and create or update merge request: %w", err)
	}

	return nil
}
