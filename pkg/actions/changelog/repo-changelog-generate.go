package changelog

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/state"
	"path/filepath"
	"time"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ChangelogGenerateStruct struct{}

// GetDetails retrieves information about the action
func (action ChangelogGenerateStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "repo-changelog-generate",
		Version:   "0.1.0",
		UsedTools: []string{},
	}
}

// Execute runs the action
func (action ChangelogGenerateStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	// retrieve commits
	commits, commitsErr := vcsrepository.FindCommitsBetweenRefs(ctx.ProjectDir, ctx.Env["NCI_COMMIT_REF_VCS"], ctx.Env["NCI_LASTRELEASE_REF_VCS"])
	if commitsErr != nil {
		log.Error().Str("from", ctx.Env["NCI_COMMIT_REF_VCS"]).Str("to", ctx.Env["NCI_LASTRELEASE_REF_VCS"]).Msg("failed to retrieve commits between refs")
	}

	// preprocess
	commits = PreprocessCommits(&config, commits)

	// analyze / grouping
	templateData := ProcessCommits(&config, commits)
	templateData.ProjectName = ctx.Env["NCI_PROJECT_NAME"]
	templateData.ProjectURL = ctx.Env["NCI_REPOSITORY_PROJECT_URL"]
	templateData.ReleaseDate = time.Now()
	templateData.Version = ctx.Env["NCI_COMMIT_REF_NAME"]

	// render all templates
	for _, templateFile := range config.Templates {
		log.Debug().Str("template", templateFile).Msg("processing template")

		content, contentErr := GetFileContent(".cid/templates", TemplateFS, templateFile)
		if contentErr != nil {
			return errors.New("failed to retrieve template content from file " + templateFile + ". " + contentErr.Error())
		}

		// render
		output, outputErr := RenderTemplate(&templateData, content)
		if outputErr != nil {
			return errors.New("failed to render template " + templateFile)
		}

		// savet to file
		filesystem.CreateDirectory(filepath.Join(ctx.Paths.Artifact, "generated", "changelog"))
		targetFile := filepath.Join(ctx.Paths.Artifact, "generated", "changelog", templateFile)
		saveErr := filesystem.SaveFileText(targetFile, output)
		if saveErr != nil {
			return errors.New("failed to save changelog file of " + templateFile + " to " + targetFile)
		}

		log.Info().Str("template", templateFile).Str("output-file", targetFile).Msg("rendered changelog template successfully")
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(ChangelogGenerateStruct{})
}
