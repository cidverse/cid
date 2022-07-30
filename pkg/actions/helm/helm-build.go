package helm

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/version"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "helm-build",
		Version:   "0.1.0",
		UsedTools: []string{"helm"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	chartDir := filepath.Join(ctx.Paths.Artifact, "helm-charts")

	// version
	ver := ""

	// helm repo add for all used repositories - https://github.com/helm/helm/issues/8036 and https://github.com/helm/helm/issues/9903
	chartFile := filepath.Join(ctx.CurrentModule.Directory, "Chart.yaml")
	chart := ParseChart(chartFile)
	if chart != nil && len(chart.Dependencies) > 0 {
		repos := make(map[string]bool)

		// collect repos
		for _, dep := range chart.Dependencies {
			repos[dep.Repository] = true
		}

		// keep version
		ver = chart.Version

		if chart.Annotations["cid.version"] == "auto" {
			ver = ctx.Env["NCI_COMMIT_REF_RELEASE"]

			// use nextrelease name as prerelease, if not valid ref is present
			if !version.IsValidSemver(ver) && len(ctx.Env["NCI_NEXTRELEASE_NAME"]) > 0 {
				ver = ctx.Env["NCI_NEXTRELEASE_NAME"]
			}
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
	log.Info().Str("path", ctx.CurrentModule.Directory).Str("version", ver).Msg("packaging chart")
	command.RunCommand("helm package "+ctx.CurrentModule.Directory+" --version "+ver+" --destination "+chartDir, ctx.Env, ctx.ProjectDir)

	// index chartDir
	command.RunCommand("helm repo index "+chartDir, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
