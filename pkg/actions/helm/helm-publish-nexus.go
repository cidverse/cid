package helm

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/actioncommon"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"path/filepath"
)

type PublishActionStruct struct{}

// GetDetails returns information about this action
func (action PublishActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "helm-publish-nexus",
		Version:   "0.1.0",
		UsedTools: []string{"helm"},
	}
}

// Check evaluates if the action should be executed or not
func (action PublishActionStruct) Check(ctx api.ActionExecutionContext) bool {
	if ctx.CurrentModule != nil && (ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemHelm) {
		return ctx.Env["NCI_COMMIT_REF_TYPE"] == "tag" && actioncommon.MapContainsKey(ctx.MachineEnv, "HELM_NEXUS_URL")
	}

	return false
}

// Execute runs the action
func (action PublishActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	// globals
	chartDir := filepath.Join(ctx.Paths.Artifact, "helm-charts")
	// config
	nexusUrl := api.GetEnvValue(ctx, "HELM_NEXUS_URL")
	nexusUsername := api.GetEnvValue(ctx, "HELM_NEXUS_USERNAME")
	nexusPassword := api.GetEnvValue(ctx, "HELM_NEXUS_PASSWORD")

	// publish
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemHelm {
		files, filesErr := filesystem.FindFilesByExtension(chartDir, []string{"tgz"})
		if filesErr != nil {
			log.Warn().Str("chart-dir", chartDir).Msg("failed to get files from chart dir")
		}

		for _, file := range files {
			log.Info().Str("nexus", nexusUrl).Str("chart", file).Msg("uploading chart to nexus")
			UploadChart(nexusUrl, nexusUsername, nexusPassword, file)
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}
