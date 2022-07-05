package container

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strings"
	"time"
)

type DockerPackageActionStruct struct{}

// GetDetails retrieves information about the action
func (action DockerPackageActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "container-package-docker",
		Version:   "0.1.0",
		UsedTools: []string{"docker"},
	}
}

// Check evaluates if the action should be executed or not
func (action DockerPackageActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return filesystem.FileExists(filepath.Join(ctx.ProjectDir, "Dockerfile")) || len(DetectAppType(ctx)) > 0
}

// Execute runs the action
func (action DockerPackageActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	dockerfile := filepath.Join(ctx.CurrentModule.Directory, "Dockerfile")
	image := getFullImage(ctx.Env["NCI_CONTAINERREGISTRY_HOST"], ctx.Env["NCI_CONTAINERREGISTRY_REPOSITORY"], ctx.Env["NCI_CONTAINERREGISTRY_TAG"])

	if filesystem.FileExists(dockerfile) {
		dockerfileContent, _ := filesystem.GetFileContent(dockerfile)

		syntax := getDockerfileSyntax(dockerfileContent)
		platforms := getDockerfileTargetPlatforms(dockerfileContent)
		targetImage := getDockerfileTargetImage(dockerfileContent)
		if len(targetImage) > 0 {
			image = targetImage
		}
		log.Info().Str("syntax", syntax).Interface("platforms", platforms).Str("image", image).Msg("building container image")

		// build args
		var buildArgs []string
		buildArgs = append(buildArgs, `docker build`)
		buildArgs = append(buildArgs, `--label "org.opencontainers.image.source=`+strings.TrimSuffix(ctx.Env["NCI_REPOSITORY_REMOTE"], ".git")+`"`)
		buildArgs = append(buildArgs, `--label "org.opencontainers.image.created=`+time.Now().Format(time.RFC3339)+`"`)
		buildArgs = append(buildArgs, `--label "org.opencontainers.image.authors="`)
		buildArgs = append(buildArgs, `--label "org.opencontainers.image.title=`+ctx.CurrentModule.Name+`"`)
		buildArgs = append(buildArgs, `-o type=oci,dest=oci.tar`)
		buildArgs = append(buildArgs, `-t `+image)
		//buildArgs = append(buildArgs, `-f Dockerfile`)
		buildArgs = append(buildArgs, ctx.CurrentModule.Directory)

		// build image
		command.RunCommand(strings.Join(buildArgs, " "), ctx.Env, ctx.ProjectDir)
	}

	// publish image
	if len(ctx.Env["NCI_CONTAINERREGISTRY_HOST"]) > 0 {
		// command.RunCommand("docker push "+image, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(DockerPackageActionStruct{})
}