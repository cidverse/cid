package appgithub

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appmergerequest"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/gosimple/slug"
	"gopkg.in/yaml.v3"
)

// GitHubWorkflowTask generates a project-specific GitHub workflow file and creates a pull request
//
// Links of interest:
// https://github.com/actions/runner-images/blob/main/images/ubuntu/Ubuntu2404-Readme.md?plain=1
func GitHubWorkflowTask(taskContext taskcommon.TaskContext) error {
	helper := simpletask.New(taskContext)
	workflowState := appconfig.NewWorkflowState()

	// read config
	conf := appconfig.Config{
		Version:      constants.Version,
		JobTimeout:   10,
		EgressPolicy: "block",
		Workflows:    appconfig.DefaultWorkflowConfig(taskContext.Repository.DefaultBranch),
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
		for pair := conf.Workflows.Newest(); pair != nil; pair = pair.Prev() {
			wfKey := pair.Key
			wfConfig := pair.Value
			slog.With("workflow", wfKey).Debug("generating workflow file")

			filteredEnvs, wfErr := appcommon.FilterVCSEnvironments(environments, wfConfig.EnvironmentPattern)
			if wfErr != nil {
				return fmt.Errorf("failed to filter workflow environments [%s]: %w", wfKey, err)
			}

			workflowTemplateData, wfErr := appconfig.GenerateWorkflowData(cid, taskContext, conf, wfKey, wfConfig, filteredEnvs, githubWorkflowDependencies, githubNetworkAllowList)
			if wfErr != nil {
				return fmt.Errorf("failed to generate workflow template [%s]: %w", wfKey, wfErr)
			}

			data, wfErr := renderWorkflow(workflowTemplateData, "wf-main.gohtml", filepath.Join(taskContext.Directory, fmt.Sprintf(".github/workflows/cid-%s.yml", slug.Make(wfKey))))
			if wfErr != nil {
				return fmt.Errorf("failed to render workflow [%s]: %w", wfKey, wfErr)
			}

			err = appconfig.PersistPlan(data.Plan, filepath.Join(taskContext.Directory, fmt.Sprintf(".github/cid/plans/%s.json", slug.Make(wfKey))))
			if err != nil {
				return fmt.Errorf("failed to persist workflow plan [%s]: %w", wfKey, err)
			}

			workflowState.Workflows.Set(wfKey, workflowTemplateData)
		}
	}

	// write workflow state
	previousState, _ := appconfig.ReadWorkflowState(filepath.Join(taskContext.Directory, ".cid", "state.json"))
	err = appconfig.WriteWorkflowState(workflowState, filepath.Join(taskContext.Directory, ".cid", "state.json"))
	if err != nil {
		return fmt.Errorf("failed to write workflow state: %w", err)
	}

	// description
	title, description, err := appmergerequest.TitleAndDescription(conf.Version, workflowState, previousState, mergeRequestFooter)
	if err != nil {
		return fmt.Errorf("failed to get merge request description: %w", err)
	}

	// hash workflow
	workflowHash, err := workflowState.Hash()
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
