package apptask

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appmergerequest"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/simpletask"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
)

type PlatformWorkflowData struct {
	Conf             appconfig.Config
	CidContext       *context.CIDContext
	Environments     map[string]appcommon.VCSEnvironment
	ProjectVariables []api.CIVariable
}

type PlatformWorkflowTaskOptions struct {
	LoadConfig func(taskContext taskcommon.TaskContext) (appconfig.Config, error) // LoadConfig defines a function that loads the configuration for the workflow task.

	// RenderWorkflow defines a function that renders the workflow file based on the provided parameters.
	RenderWorkflow func(
		workflowState *appconfig.WorkflowState,
		data PlatformWorkflowData,
		template string,
		targetPath string,
	) (string, error)

	WorkflowStatePath  string // Path where workflow state JSON will be stored.
	MergeRequestFooter string // MergeRequestFooter defines the footer of the merge request description, e.g. "This MR was auto-generated by CID."
}

func WorkflowTaskData(taskContext taskcommon.TaskContext, opts PlatformWorkflowTaskOptions) (PlatformWorkflowData, error) {
	// load config
	conf, err := opts.LoadConfig(taskContext)
	if err != nil {
		return PlatformWorkflowData{}, err
	}

	// app context
	cid, err := context.NewAppContextFromDir(taskContext.Directory, taskContext.Directory)
	if err != nil {
		return PlatformWorkflowData{}, err
	}

	// vars
	projectVariables, err := taskContext.Platform.Variables(taskContext.Repository) // TODO: unused?
	if err != nil {
		return PlatformWorkflowData{}, fmt.Errorf("failed to get variables: %w", err)
	}

	// env
	envs, err := taskContext.Platform.Environments(taskContext.Repository)
	if err != nil {
		return PlatformWorkflowData{}, fmt.Errorf("failed to get environments: %w", err)
	}
	environments := make(map[string]appcommon.VCSEnvironment, len(envs))
	for _, e := range envs {
		vars, err := taskContext.Platform.EnvironmentVariables(taskContext.Repository, e.Name)
		if err != nil {
			return PlatformWorkflowData{}, fmt.Errorf("failed to get environment variables: %w", err)
		}

		environments[e.Name] = appcommon.VCSEnvironment{
			Env:  e,
			Vars: vars,
		}
	}

	return PlatformWorkflowData{
		Conf:             conf,
		CidContext:       cid,
		Environments:     environments,
		ProjectVariables: projectVariables,
	}, nil
}

type WorkflowTaskResult struct {
	WorkflowContent map[string]string        // WorkflowContent is a map with the rendered content of each generated file.
	WorkflowState   *appconfig.WorkflowState // WorkflowState is the state of the workflow after processing.
}

func WorkflowTask(taskContext taskcommon.TaskContext, opts PlatformWorkflowTaskOptions, dryRun bool) (WorkflowTaskResult, error) {
	helper := simpletask.New(taskContext)
	workflowState := appconfig.NewWorkflowState()
	renderedFiles := make(map[string]string)
	stateFile := opts.WorkflowStatePath

	// load config
	conf, err := opts.LoadConfig(taskContext)
	if err != nil {
		return WorkflowTaskResult{}, err
	}

	// clone repository
	slog.With("dir", taskContext.Directory).With("clone-uri", taskContext.Repository.CloneSSH).Debug("cloning repository")
	err = helper.Clone()
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to clone repository: %w", err)
	}

	// workflow data
	data, err := WorkflowTaskData(taskContext, opts)
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to prepare workflow task data: %w", err)
	}

	// create and checkout new branch
	branch := fmt.Sprintf("%s-%s", appcommon.BranchName, conf.Version)
	err = helper.CreateBranch(branch)
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to create branch: %w", err)
	}

	// render workflows
	if conf.Workflows != nil {
		content, err := opts.RenderWorkflow(workflowState, data, "wf-main.gohtml", filepath.Join(taskContext.Directory, ".gitlab-ci.yml"))
		if err != nil {
			return WorkflowTaskResult{}, fmt.Errorf("failed to render workflow: %w", err)
		}

		renderedFiles[".gitlab-ci.yml"] = content
	}

	// check if dry run mode is enabled
	if dryRun {
		// render state file
		stateJson, err := appconfig.WorkflowStateJSON(workflowState)
		if err != nil {
			return WorkflowTaskResult{}, fmt.Errorf("failed to marshal workflow state to JSON: %w", err)
		}
		relativeStateFile := strings.TrimPrefix(stateFile, taskContext.Directory+"/")
		renderedFiles[relativeStateFile] = stateJson

		// log dry run mode
		slog.Info("Dry run mode enabled, skipping commit and merge request creation.")
		return WorkflowTaskResult{WorkflowState: workflowState, WorkflowContent: renderedFiles}, nil
	}

	// write workflow state
	previousState, _ := appconfig.ReadWorkflowState(stateFile)
	err = appconfig.WriteWorkflowState(workflowState, stateFile)
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to write workflow state: %w", err)
	}

	// description
	title, description, err := appmergerequest.TitleAndDescription(conf.Version, *workflowState, previousState, opts.MergeRequestFooter)
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to get merge request description: %w", err)
	}

	// hash workflow
	workflowHash, err := workflowState.Hash()
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to hash workflow contents: %w", err)
	}

	// commit push and create or update merge request
	err = helper.CommitPushAndMergeRequest(title, title, description, fmt.Sprintf("%s-%s", appcommon.MergeRequestId, workflowHash))
	if err != nil {
		return WorkflowTaskResult{}, fmt.Errorf("failed to commit push and create or update merge request: %w", err)
	}

	return WorkflowTaskResult{WorkflowState: workflowState, WorkflowContent: renderedFiles}, nil
}
