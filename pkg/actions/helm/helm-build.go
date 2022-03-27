package helm

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"path/filepath"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "helm-build",
		Version:   "0.1.0",
		UsedTools: []string{"helm"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemHelm
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	chartDir := filepath.Join(ctx.Paths.Artifact, "helm-charts")

	// rebuild the charts/ directory based on the Chart.lock file
	command.RunCommand("helm dependency build "+ctx.CurrentModule.Directory, ctx.Env, ctx.ProjectDir)

	// package charts
	command.RunCommand("helm package "+ctx.CurrentModule.Directory+" --destination "+chartDir, ctx.Env, ctx.ProjectDir)

	// index chartDir
	command.RunCommand("helm repo index "+chartDir, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
