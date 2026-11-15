package appgitlab

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/apptask"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
)

// GitLabWorkflowTask generates a project-specific GitLab workflow file and creates a pull request
//
// Links of interest:
// https://gitlab.com/gitlab-org/gitlab/-/tree/master/doc/ci/runners/hosted_runners?ref_type=heads for all available runner tags
// GitLabWorkflowTask renders .gitlab-ci.yml and, unless opts.DryRun, commits it and opens a merge request.
func GitLabWorkflowTask(taskContext taskcommon.TaskContext, opts TaskOptions) (apptask.WorkflowTaskResult, error) {
	return apptask.WorkflowTask(taskContext, apptask.PlatformWorkflowTaskOptions{
		Config: appconfig.Config{
			Version:          constants.Version,
			VersionHash:      constants.BinaryHash,
			JobTimeout:       20,
			RunnerTags:       BaseRunnerTags(opts.RunnerTags), // project variables are merged in by apptask.WorkflowTask
			EgressPolicy:     "block",
			ContainerRuntime: "podman",
			Workflows:        appconfig.DefaultWorkflowConfig(taskContext.Repository.DefaultBranch, taskContext.Repository.Branches),
		},
		RenderWorkflow: func(workflowState *appconfig.WorkflowState, data apptask.PlatformWorkflowData, template string, targetPath string) (string, error) {
			var workflowTemplateData []appconfig.WorkflowData

			for pair := data.Conf.Workflows.Newest(); pair != nil; pair = pair.Prev() {
				wfKey := pair.Key
				wfConfig := pair.Value

				filteredEnvs, wfErr := appcommon.FilterVCSEnvironments(data.Environments, wfConfig.EnvironmentPattern)
				if wfErr != nil {
					return "", fmt.Errorf("failed to filter workflow environments [%s]: %w", wfKey, wfErr)
				}

				wtd, wfErr := appconfig.GenerateWorkflowData(data.CidContext, taskContext, data.Conf, wfKey, wfConfig, data.ProjectVariables, filteredEnvs, gitlabWorkflowDependencies, gitlabNetworkAllowList)
				if wfErr != nil {
					return "", fmt.Errorf("failed to generate workflow template [%s]: %w", wfKey, wfErr)
				}

				for i := range wtd.Plan.Steps { // TD-001: add gitlab-sarif-converter to steps that produce SARIF reports, due to automatic report conversion for GitLab
					if wtd.Plan.Steps[i].Outputs.ContainsArtifactWithTypeAndFormat("report", "sarif") {
						wtd.Plan.Steps[i].Access.Executables = append(wtd.Plan.Steps[i].Access.Executables, actionsdk.ActionAccessExecutable{Name: "gitlab-sarif-converter"})
						wtd.Plan.Steps[i].Outputs.Artifacts = append(wtd.Plan.Steps[i].Outputs.Artifacts, actionsdk.ActionArtifactType{Type: "report", Format: "gl-codequality"})
					}
				}

				workflowTemplateData = append(workflowTemplateData, wtd)
				workflowState.Workflows.Set(wfKey, &wtd)
			}

			// render workflow
			wfResult, wfErr := renderWorkflow(workflowTemplateData, data.Conf.RunnerTags, "wf-main.gohtml", filepath.Join(taskContext.Directory, ".gitlab-ci.yml"), RenderOptions{SelfContained: opts.SelfContained})
			if wfErr != nil {
				return "", fmt.Errorf("failed to render [gitlab-ci.yml]: %w", wfErr)
			}

			return wfResult.WorkflowContent, nil
		},
		AutoMergeAfter:     opts.AutoMergeAfter,
		WorkflowStatePath:  filepath.Join(taskContext.Directory, ".cid", "state-gitlab.json"),
		MergeRequestFooter: mergeRequestFooter,
	}, opts.DryRun)
}

// TaskOptions controls GitLabWorkflowTask.
type TaskOptions struct {
	// DryRun renders without committing or opening a merge request.
	DryRun bool
	// SelfContained renders a pipeline without .cid/ file dependencies (inlined install script, embedded plans), see RenderOptions.
	SelfContained bool
	// RunnerTags configures the runner tags (default / per stage), applied on top of DefaultRunnerTags; project CI/CD variables override them.
	RunnerTags appconfig.RunnerTags
	// AutoMergeAfter enables auto-merging cid's merge request once it is this old (0 disables), see appmergerequest.AutoMerge.
	// Project CI/CD variables override it.
	AutoMergeAfter time.Duration
}

// DefaultRunnerTags are the default runner tags of GitLab jobs (gitlab.com hosted runners).
var DefaultRunnerTags = []string{"saas-linux-small-amd64"}

// BaseRunnerTags returns the GitLab defaults overridden by configured.
func BaseRunnerTags(configured appconfig.RunnerTags) appconfig.RunnerTags {
	return appconfig.RunnerTags{Default: DefaultRunnerTags}.Merge(configured)
}
