package container

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

type PackageActionStruct struct{}

// GetDetails retrieves information about the action
func (action PackageActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "package",
		Name:      "container-package",
		Version:   "0.1.0",
		UsedTools: []string{"docker"},
	}
}

// Check evaluates if the action should be executed or not
func (action PackageActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return len(DetectAppType(ctx)) > 0
}

// Execute runs the action
func (action PackageActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	dockerfile := ctx.ProjectDir + `/Dockerfile`

	// auto detect a usable dockerfile
	appType := DetectAppType(ctx)
	if appType == "jar" {
		dockerfileContent, dockerfileContentErr := api.GetFileContentFromEmbedFS(DockerfileFS, "dockerfiles/Java15.Dockerfile")
		if dockerfileContentErr != nil {
			log.Fatal().Err(dockerfileContentErr).Msg("failed to get dockerfile from resources.")
		}

		createFileErr := filesystem.CreateFileWithContent(dockerfile, dockerfileContent)
		if createFileErr != nil {
			log.Fatal().Str("file", dockerfile).Msg("failed to create temporary dockerfile")
		}
	}

	// run build
	command.RunCommand(`docker build --no-cache -t `+ctx.Env["NCI_CONTAINERREGISTRY_REPOSITORY"]+":"+ctx.Env["NCI_COMMIT_REF_RELEASE"]+` `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)

	// remove dockerfile
	_ = filesystem.RemoveFile(dockerfile)
	return nil
}

func init() {
	api.RegisterBuiltinAction(PackageActionStruct{})
}
