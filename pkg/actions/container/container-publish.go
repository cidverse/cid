package container

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/google/uuid"
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

// Execute runs the action
func (action PublishActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	// find registry auth file
	authFile := getFirstExistingFile([]string{
		"/var/tmp/containers-user-" + ctx.CurrentUser.Uid + "/containers/containers/auth.json",
	})
	ctx.Env["REGISTRY_AUTH_FILE"] = authFile

	// target image reference
	imageRefFile := filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), "image.txt")
	imageRef, imageRefErr := filesystem.GetFileContent(imageRefFile)
	if imageRefErr != nil {
		return errors.New("failed to parse image reference from " + imageRefFile)
	}

	// dockerhub still has some issues with the oci format
	format := "oci"
	if strings.HasPrefix(imageRef, "docker.io/") {
		format = "v2s2"
	}

	// for each container archive
	var files []string
	var _ = filepath.WalkDir(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".tar") {
			files = append(files, path)
		}

		return nil
	})
	if len(files) == 0 {
		log.Error().Str("path", ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image")).Msg("no candidates for image publication found!")
		return nil
	}

	// prepare manifest
	manifestName := strings.Replace(uuid.NewString(), "-", "", -1)
	log.Info().Str("manifest", manifestName).Str("ref", imageRef).Int("files", len(files)).Msg("publishing image using manifest ...")
	command.RunCommand("buildah manifest create "+manifestName, ctx.Env, ctx.ProjectDir)

	// add images to manifest
	for _, file := range files {
		log.Info().Str("manifest", manifestName).Str("ref", imageRef).Str("file", file).Msg("add image to manifest")
		command.RunCommand("buildah manifest add "+manifestName+" oci-archive:"+file, ctx.Env, ctx.ProjectDir)
	}

	// print manifest
	command.RunCommand("buildah manifest inspect "+manifestName, ctx.Env, ctx.ProjectDir)

	// publish manifest to registry
	log.Info().Str("manifest", manifestName).Str("ref", imageRef).Msg("uploading manifest ...")
	command.RunCommand("buildah manifest push --all --format "+format+" "+manifestName+" docker://"+imageRef, ctx.Env, ctx.ProjectDir) // format: v2s2 or oci

	return nil
}

func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}
