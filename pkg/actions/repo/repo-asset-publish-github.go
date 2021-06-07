package repo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strings"
)

type AssetPublishGitHubStruct struct {}

// GetDetails returns information about this action
func (action AssetPublishGitHubStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "publish",
		Name: "repo-asset-publish-github",
		Version: "0.1.0",
		UsedTools: []string{"gh"},
	}
}

// Check if this package can handle the current environment
func (action AssetPublishGitHubStruct) Check(ctx api.ActionExecutionContext) bool {
	if len(ctx.MachineEnv["GITHUB_TOKEN"]) > 0 && strings.HasPrefix(ctx.Env["NCI_REPOSITORY_REMOTE"], "https://github.com") {
		ctx.Env["GITHUB_TOKEN"] = ctx.MachineEnv["GITHUB_TOKEN"]

		return ctx.Env["NCI_COMMIT_REF_TYPE"]  == "tag"
	}

	return false
}

// Check if this package can handle the current environment
func (action AssetPublishGitHubStruct) Execute(ctx api.ActionExecutionContext) {
	tagName := ctx.Env["NCI_COMMIT_REF_NAME"]

	// create github release
	var opts []string

	// title
	opts = append(opts, `--title "`+tagName+`"`)

	// prerelease?
	if !api.IsVersionStable(tagName) {
		opts = append(opts, "--prerelease")
	}

	// TODO: changelog
	opts = append(opts, `--notes "..."`) // -F changelog.md

	log.Info().Str("release", tagName).Msg("creating github release ...")
	command.RunCommand(`gh release create `+tagName+` `+strings.Join(opts, " "), ctx.Env, ctx.ProjectDir)

	// upload artifacts
	if filesystem.DirectoryExists(filepath.Join(ctx.ProjectDir, "dist", "bin")) {
		files, filesErr := filesystem.FindFilesInDirectory(filepath.Join(ctx.ProjectDir, "dist", "bin"), "")
		if filesErr != nil {
			// err
		} else {
			for _, file := range files {
				if filesystem.FileExists(file) {
					opts = append(opts, file)
					log.Info().Str("file", file).Msg("uploading github release asset ...")
					command.RunCommand(`gh release upload `+tagName+` `+cihelper.ToUnixPath(file), ctx.Env, ctx.ProjectDir)
				}
			}
		}
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(AssetPublishGitHubStruct{})
}