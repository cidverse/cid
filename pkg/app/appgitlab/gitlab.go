package appgitlab

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
	"github.com/cidverse/cidverseutils/hash"
	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/elliotchance/orderedmap/v3"
	"gopkg.in/yaml.v3"
)

func GitLabWorkflowTask(taskContext taskcommon.TaskContext) error {
	helper := simpletask.New(taskContext)
	var generatedContent string

	// read config
	workflowMap := orderedmap.NewOrderedMap[string, appconfig.WorkflowConfig]()
	workflowMap.Set("Main", appconfig.WorkflowConfig{
		Type:                "main",
		TriggerManual:       true,
		TriggerPush:         true,
		TriggerPushBranches: []string{taskContext.Repository.DefaultBranch},
	})
	workflowMap.Set("Release", appconfig.WorkflowConfig{
		Type:               "release",
		TriggerManual:      true,
		TriggerPush:        true,
		TriggerPushTags:    []string{"v*.*.*"},
		EnvironmentPattern: "release-.*",
	})
	workflowMap.Set("Pull Request", appconfig.WorkflowConfig{
		Type:               "pull-request",
		TriggerPullRequest: true,
		EnvironmentPattern: "pr-.*",
	})
	workflowMap.Set("Nightly", appconfig.WorkflowConfig{
		Type:                "nightly",
		TriggerManual:       true,
		TriggerSchedule:     true,
		TriggerScheduleCron: "@daily",
		EnvironmentPattern:  "nightly-.*",
	})
	conf := appconfig.Config{
		Version:      constants.Version,
		JobTimeout:   10,
		EgressPolicy: "block",
		Workflows:    workflowMap,
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

	// env
	envs, err := taskContext.Platform.Environments(taskContext.Repository)
	if err != nil {
		return fmt.Errorf("failed to get environments: %w", err)
	}

	environments := make(map[string]appcommon.VCSEnvironment, len(envs))
	for _, e := range envs {
		// fetch environment variables
		vars, err := taskContext.Platform.EnvironmentVariables(taskContext.Repository, e.Name)
		if err != nil {
			return fmt.Errorf("failed to get environment variables: %w", err)
		}

		environments[e.Name] = appcommon.VCSEnvironment{
			Env:  e,
			Vars: vars,
		}
	}

	// workflows
	if conf.Workflows != nil {
		var workflowTemplateData []appconfig.WorkflowData

		for wfKey, wfConfig := range conf.Workflows.AllFromFront() {
			filteredEnvs, wfErr := appcommon.FilterVCSEnvironments(environments, wfConfig.EnvironmentPattern)
			if wfErr != nil {
				return fmt.Errorf("failed to filter workflow environments [%s]: %w", wfKey, err)
			}

			wtd, wfErr := appconfig.GenerateWorkflowData(cid, taskContext, conf, wfKey, wfConfig, filteredEnvs, nil, gitlabNetworkAllowList)
			if wfErr != nil {
				return fmt.Errorf("failed to generate workflow template [%s]: %w", wfKey, wfErr)
			}

			workflowTemplateData = append(workflowTemplateData, wtd)
		}

		// render workflow
		_, wfErr := renderWorkflow(workflowTemplateData, "wf-main.gohtml", filepath.Join(taskContext.Directory, ".gitlab-ci.yml"))
		if wfErr != nil {
			return fmt.Errorf("failed to render [gitlab-ci.yml]: %w", wfErr)
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
