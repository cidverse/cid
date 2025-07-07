package appgitlab

import (
	"fmt"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/apptask"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
)

// GitLabWorkflowTask generates a project-specific GitLab workflow file and creates a pull request
//
// Links of interest:
// https://gitlab.com/gitlab-org/gitlab/-/tree/master/doc/ci/runners/hosted_runners?ref_type=heads for all available runner tags
func GitLabWorkflowTask(taskContext taskcommon.TaskContext, dryRun bool) (apptask.WorkflowTaskResult, error) {
	return apptask.WorkflowTask(taskContext, apptask.PlatformWorkflowTaskOptions{
		LoadConfig: func(taskContext taskcommon.TaskContext) (appconfig.Config, error) {
			return appconfig.Config{
				Version:          constants.Version,
				JobTimeout:       10,
				RunnerTags:       []string{"saas-linux-small-amd64"},
				EgressPolicy:     "block",
				ContainerRuntime: "podman",
				Workflows:        appconfig.DefaultWorkflowConfig(taskContext.Repository.DefaultBranch),
			}, nil
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
						wtd.Plan.Steps[i].Access.Executables = append(wtd.Plan.Steps[i].Access.Executables, catalog.ActionAccessExecutable{Name: "gitlab-sarif-converter"})
						wtd.Plan.Steps[i].Outputs.Artifacts = append(wtd.Plan.Steps[i].Outputs.Artifacts, catalog.ActionArtifactType{Type: "report", Format: "gl-codequality"})
					}
				}

				workflowTemplateData = append(workflowTemplateData, wtd)
				workflowState.Workflows.Set(wfKey, &wtd)
			}

			// render workflow
			wfResult, wfErr := renderWorkflow(workflowTemplateData, data.Conf.RunnerTags, "wf-main.gohtml", filepath.Join(taskContext.Directory, ".gitlab-ci.yml"))
			if wfErr != nil {
				return "", fmt.Errorf("failed to render [gitlab-ci.yml]: %w", wfErr)
			}

			return wfResult.WorkflowContent, nil
		},
		WorkflowStatePath:  filepath.Join(taskContext.Directory, ".cid", "state-gitlab.json"),
		MergeRequestFooter: mergeRequestFooter,
	}, dryRun)
}
