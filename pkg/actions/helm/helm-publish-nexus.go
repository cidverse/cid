package helm

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
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
	var missingRequirements []api.MissingRequirement

	if ctx.CurrentModule != nil {
		if ctx.CurrentModule.BuildSystem != analyzerapi.BuildSystemHelm {
			missingRequirements = append(missingRequirements, api.MissingRequirement{Message: "module build system is not helm"})
		} else {
			if !funk.Contains(ctx.MachineEnv, "HELM_NEXUS_URL") {
				missingRequirements = append(missingRequirements, api.MissingRequirement{Message: "HELM_NEXUS_URL is not set"})
			}
		}
	} else {
		missingRequirements = append(missingRequirements, api.MissingRequirement{Message: "no module context present"})
	}

	for _, r := range missingRequirements {
		log.Warn().Str("condition", r.Message).Msg("action requirements not fulfilled")
	}

	return len(missingRequirements) == 0
}

// Execute runs the action
func (action PublishActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	// globals
	chartArtifactDir := filepath.Join(ctx.Paths.Artifact, "helm-charts")
	// config
	nexusUrl := api.GetEnvValue(ctx, "HELM_NEXUS_URL")
	nexusRepo := api.GetEnvValue(ctx, "HELM_NEXUS_REPOSITORY")
	nexusUsername := api.GetEnvValue(ctx, "HELM_NEXUS_USERNAME")
	nexusPassword := api.GetEnvValue(ctx, "HELM_NEXUS_PASSWORD")

	// publish
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemHelm {
		files, filesErr := filesystem.FindFilesByExtension(chartArtifactDir, []string{"tgz"})
		if filesErr != nil {
			log.Warn().Str("chart-dir", chartArtifactDir).Msg("failed to get files from chart dir")
		}

		for _, file := range files {
			endpoint := nexusUrl+"/service/rest/v1/components?repository="+nexusRepo
			log.Info().Str("nexus", endpoint).Str("chart", file).Msg("uploading chart to nexus")
			status, response := UploadChart(endpoint, nexusUsername, nexusPassword, file)
			if (status < 200 || status >= 300) {
				log.Warn().Str("chart", file).Int("status", status).Str("response", string(response)).Msg("chart upload failed")
			}
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}
