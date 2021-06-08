package changelog

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"time"
)

type ChangelogGenerateStruct struct{}

func (action ChangelogGenerateStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "publish",
		Name:      "repo-changelog-generate",
		Version:   "0.1.0",
		UsedTools: []string{},
	}
}

func (action ChangelogGenerateStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.Env["NCI_COMMIT_REF_TYPE"] == "tag"
}

func (action ChangelogGenerateStruct) Execute(ctx api.ActionExecutionContext) {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		log.Error().Err(configParseErr).Str("action", "repo-changelog-generate").Msg("failed to parse action configuration")
		return
	}

	// retrieve commits
	commits, commitsErr := vcsrepository.FindCommitsBetweenRefs(ctx.ProjectDir, ctx.Env["NCI_COMMIT_REF_VCS"], ctx.Env["NCI_LASTRELEASE_REF_VCS"])
	if commitsErr != nil {
		log.Error().Str("from", ctx.Env["NCI_COMMIT_REF_VCS"]).Str("to", ctx.Env["NCI_LASTRELEASE_REF_VCS"]).Msg("failed to retrieve commits between refs")
	}

	// preprocess
	commits = PreprocessCommits(config, commits)

	// analyse / grouping
	templateData := ProcessCommits(config, commits)
	templateData.ProjectName = ctx.Env["NCI_PROJECT_NAME"]
	templateData.ProjectUrl = ctx.Env["NCI_REPOSITORY_PROJECT_URL"]
	templateData.ReleaseDate = time.Now()
	templateData.Version = ctx.Env["NCI_COMMIT_REF_NAME"]

	// render all templates
	for _, templateFile := range config.Templates {
		log.Debug().Str("template", templateFile).Msg("processing template")

		content, contentErr := GetFileContent(".cid/templates", TemplateFS, templateFile)
		if contentErr != nil {
			log.Error().Err(contentErr).Str("template", templateFile).Msg("failed to get template content")
			return
		}

		// render
		output, outputErr := RenderTemplate(templateData, content)
		if outputErr != nil {
			log.Error().Err(outputErr).Str("template", templateFile).Msg("failed to render template")
			return
		}

		// save into tmp file
		targetPath := filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, "changelog")
		targetFile := filepath.Join(targetPath, templateFile)
		filesystem.CreateDirectory(targetPath)
		_ = filesystem.RemoveFile(targetFile)
		saveErr := filesystem.SaveFileContent(targetFile, output)
		if saveErr != nil {
			log.Error().Err(saveErr).Str("template", templateFile).Str("output-file", targetPath).Msg("failed to save file")
			return
		}

		log.Info().Str("template", templateFile).Str("output-file", targetPath).Msg("rendered changelog template successfully")
	}
}

func init() {
	api.RegisterBuiltinAction(ChangelogGenerateStruct{})
}
