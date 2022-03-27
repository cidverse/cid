package repo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strings"
)

type AssetPublishGitHubStruct struct{}

// GetDetails retrieves information about the action
func (action AssetPublishGitHubStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "repo-release-github",
		Version:   "0.1.0",
		UsedTools: []string{"gh"},
	}
}

// Check evaluates if the action should be executed or not
func (action AssetPublishGitHubStruct) Check(ctx api.ActionExecutionContext) bool {
	if len(ctx.MachineEnv["GITHUB_TOKEN"]) > 0 && strings.HasPrefix(ctx.Env["NCI_REPOSITORY_REMOTE"], "https://github.com") {
		return ctx.Env["NCI_COMMIT_REF_TYPE"] == "tag"
	}

	return false
}

// Execute runs the action
func (action AssetPublishGitHubStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	// context
	ctx.Env["GITHUB_TOKEN"] = ctx.MachineEnv["GITHUB_TOKEN"]

	// input
	tagName := ctx.Env["NCI_COMMIT_REF_NAME"]

	// create github release
	var opts []string

	// title
	opts = append(opts, `--title "`+tagName+`"`)

	// prerelease?
	if !api.IsVersionStable(tagName) {
		opts = append(opts, "--prerelease")
	}

	// use generated changelog
	ghChangelogFile := filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, "changelog", "github-release.tmpl")
	if filesystem.FileExists(ghChangelogFile) {
		opts = append(opts, "-F "+ghChangelogFile)
	} else {
		opts = append(opts, `--notes "..."`)
	}

	log.Info().Str("release", tagName).Msg("creating github release ...")
	command.RunCommand(`gh release create `+tagName+` `+strings.Join(opts, " "), ctx.Env, ctx.ProjectDir)

	// upload artifacts
	if filesystem.DirectoryExists(filepath.Join(ctx.ProjectDir, "dist", "bin")) {
		files, filesErr := filesystem.FindFilesByExtension(filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, "bin"), nil)
		if filesErr == nil {
			if len(files) > 0 {
				for _, file := range files {
					if filesystem.FileExists(file) {
						log.Info().Str("file", file).Msg("uploading github release asset ...")
						command.RunCommand(`gh release upload `+tagName+` `+file, ctx.Env, ctx.ProjectDir)
					}
				}
			}
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(AssetPublishGitHubStruct{})
}
