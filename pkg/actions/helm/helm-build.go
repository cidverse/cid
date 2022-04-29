package helm

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
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

	// version
	version := ctx.Env["NCI_COMMIT_REF_RELEASE"]

	// helm repo add for all used repositories - https://github.com/helm/helm/issues/8036 and https://github.com/helm/helm/issues/9903
	chartFile := filepath.Join(ctx.CurrentModule.Directory, "Chart.yaml")
	chart := ParseChart(chartFile)
	if chart != nil && len(chart.Dependencies) > 0 {
		repos := make(map[string]bool)

		// version (set automatically if chart version is 0.0.0)
		if chart.Version != "0.0.0" {
			version = chart.Version
		}

		// collect repos
		for _, dep := range chart.Dependencies {
			repos[dep.Repository] = true
		}

		// add repos
		for repo := range repos {
			log.Debug().Str("repo", repo).Msg("add helm repo")
			command.RunCommand("helm repo add "+slug.Make(repo)+" "+repo, ctx.Env, ctx.ProjectDir)
		}
	}

	// rebuild the charts/ directory based on the Chart.lock file
	command.RunCommand("helm dependency build "+ctx.CurrentModule.Directory+" --verify", ctx.Env, ctx.ProjectDir)

	// package charts
	log.Info().Str("path", ctx.CurrentModule.Directory).Str("version", version).Msg("packaging chart")
	command.RunCommand("helm package "+ctx.CurrentModule.Directory+" --version "+version+" --destination "+chartDir, ctx.Env, ctx.ProjectDir)

	// index chartDir
	command.RunCommand("helm repo index "+chartDir, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
