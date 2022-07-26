package container

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"io/fs"
	"path/filepath"
	"strings"
)

type PublishActionStruct struct{}

// GetDetails retrieves information about the action
func (action PublishActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "container-publish",
		Version:   "0.1.0",
		UsedTools: []string{"buildah"},
	}
}

// Check evaluates if the action should be executed or not
func (action PublishActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action PublishActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// target image reference
	imageRef, imageRefErr := filesystem.GetFileContent(filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), "image.txt"))
	if imageRefErr != nil {
		return errors.New("failed to parse image reference from " + filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), "image.txt"))
	}

	// for each container archive
	var files []string
	var _ = filepath.WalkDir(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".tar") {
			files = append(files, path)
		}

		return nil
	})

	// prepare manifest
	manifestName := ctx.CurrentModule.Slug
	log.Info().Str("manifest", manifestName).Str("ref", imageRef).Int("files", len(files)).Msg("publishing image using manifest ...")
	_, _ = command.RunCommandAndGetOutput("buildah manifest rm "+manifestName, ctx.Env, ctx.ProjectDir)
	command.RunCommand("buildah manifest create "+manifestName, ctx.Env, ctx.ProjectDir)

	// add images to manifest
	for _, file := range files {
		log.Info().Str("manifest", manifestName).Str("ref", imageRef).Str("file", file).Msg("add image to manifest")
		command.RunCommand("buildah manifest add "+manifestName+" oci-archive:"+file, ctx.Env, ctx.ProjectDir)
	}

	// publish manifest to registry
	log.Info().Str("manifest", manifestName).Str("ref", imageRef).Msg("uploading manifest ...")
	command.RunCommand("buildah manifest push --all -f oci "+manifestName+" docker://"+imageRef, ctx.Env, ctx.ProjectDir) // format: v2s2 or oci

	return nil
}

func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}
