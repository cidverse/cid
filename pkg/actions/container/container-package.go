package container

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strings"
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
	return filesystem.FileExists(filepath.Join(ctx.ProjectDir, "Dockerfile")) || len(DetectAppType(ctx)) > 0
}

// Execute runs the action
func (action PackageActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	dockerfile := filepath.Join(ctx.ProjectDir, "Dockerfile")
	image := getFullImage(ctx.Env["NCI_CONTAINERREGISTRY_HOST"], ctx.Env["NCI_CONTAINERREGISTRY_REPOSITORY"], ctx.Env["NCI_CONTAINERREGISTRY_TAG"])
	createdTmpDockerfile := false

	log.Info().Str("image", image).Msg("building container image")

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
		createdTmpDockerfile = true
	}

	// build args
	var buildArgs []string
	buildArgs = append(buildArgs, `docker build`)
	buildArgs = append(buildArgs, `--label "org.opencontainers.image.source=`+strings.TrimSuffix(ctx.Env["NCI_REPOSITORY_REMOTE"], ".git")+`"`)
	buildArgs = append(buildArgs, `-t ` + image)
	buildArgs = append(buildArgs, ctx.ProjectDir)

	// build image
	command.RunCommand(strings.Join(buildArgs, " "), ctx.Env, ctx.ProjectDir)

	// publish image
	if len(ctx.Env["NCI_CONTAINERREGISTRY_HOST"]) > 0 {
		command.RunCommand("docker push "+image, ctx.Env, ctx.ProjectDir)
	}

	// remove dockerfile
	if createdTmpDockerfile {
		_ = filesystem.RemoveFile(dockerfile)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(PackageActionStruct{})
}
