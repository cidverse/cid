package container

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/state"
	"path/filepath"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type BuildahPackageActionStruct struct{}

type BuildahConfig struct {
	NoCache   bool `yaml:"no-cache" required:"true" default:"true"`
	Squash    bool `yaml:"squash" required:"true" default:"true"`
	Rebuild   bool `yaml:"rebuild" required:"true" default:"true"`
	Platforms []Platform
}

// GetDetails retrieves information about the action
func (action BuildahPackageActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "container-package-buildah",
		Version:   "0.1.0",
		UsedTools: []string{"buildah"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildahPackageActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildahPackageActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	var config BuildahConfig
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	imageReference := getFullImage(ctx.Env["NCI_CONTAINERREGISTRY_HOST"], ctx.Env["NCI_CONTAINERREGISTRY_REPOSITORY"], ctx.Env["NCI_CONTAINERREGISTRY_TAG"])
	for _, discovery := range ctx.CurrentModule.Discovery {
		containerFile := strings.TrimPrefix(discovery, "file~")
		containerFileContent, _ := filesystem.GetFileContent(containerFile)

		if ctx.CurrentModule.BuildSystemSyntax == analyzerapi.ContainerFile {
			syntax := getDockerfileSyntax(containerFileContent)
			platforms := getDockerfileTargetPlatforms(containerFileContent)
			imageReference = getDockerfileTargetImage(containerFileContent, imageReference)

			// skip image generation, if image is present in remote registry
			if !config.Rebuild {
				_, remoteImageErr := LoadRemoteImageInformation(imageReference)
				if remoteImageErr == nil {
					log.Info().Str("syntax", syntax).Interface("platforms", platforms).Str("image", imageReference).Str("cause", "present_in_remote").Msg("skipping container image build")
					return nil
				}
			}

			// build each image and add to manifest
			for _, platform := range platforms {
				containerArchiveFile := filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), platform.Platform("_")+".tar")
				log.Info().Str("syntax", syntax).Interface("platform", platform.Platform("/")).Str("image", imageReference).Msg("building container image")

				var buildArgs []string
				buildArgs = append(buildArgs, `buildah bud`)
				buildArgs = append(buildArgs, `--platform `+platform.Platform("/"))
				buildArgs = append(buildArgs, `-f `+filepath.Base(containerFile))
				buildArgs = append(buildArgs, `-t `+"oci-archive:"+containerArchiveFile)

				// options
				if config.NoCache {
					buildArgs = append(buildArgs, `--no-cache`)
				}
				if config.Squash {
					buildArgs = append(buildArgs, `--squash`) // squash, excluding the base layer
				}

				// download cache
				downloadCache := ctx.Paths.NamedCache("buildah-download/" + platform.OS + "-" + platform.Arch)
				log.Debug().Str("source", downloadCache).Msg("mounting external cache for /cache")
				buildArgs = append(buildArgs, `-v `+downloadCache+`:/cache`)

				// labels (oci annotations: https://github.com/opencontainers/image-spec/blob/main/annotations.md)
				buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.source=`+strings.TrimSuffix(ctx.Env["NCI_REPOSITORY_REMOTE"], ".git")+`"`)
				buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.created=`+time.Now().Format(time.RFC3339)+`"`)
				buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.authors="`)
				buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.title=`+ctx.CurrentModule.Name+`"`)
				buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.description="`)

				// dynamic build-args
				if strings.Contains(containerFileContent, "ARG TARGETPLATFORM") {
					buildArgs = append(buildArgs, `--build-arg TARGETPLATFORM=`+platform.Platform("/"))
				}
				if strings.Contains(containerFileContent, "ARG TARGETOS") {
					buildArgs = append(buildArgs, `--build-arg TARGETOS=`+platform.OS)
				}
				if strings.Contains(containerFileContent, "ARG TARGETARCH") {
					buildArgs = append(buildArgs, `--build-arg TARGETARCH=`+platform.Arch)
				}
				if strings.Contains(containerFileContent, "ARG TARGETVARIANT") {
					buildArgs = append(buildArgs, `--build-arg TARGETVARIANT=`+platform.Variant)
				}

				buildArgs = append(buildArgs, ctx.CurrentModule.Directory)
				command.RunCommand(strings.Join(buildArgs, " "), ctx.Env, ctx.ProjectDir)
			}
		} else if ctx.CurrentModule.BuildSystemSyntax == analyzerapi.ContainerBuildahScript {
			platforms := getDockerfileTargetPlatforms(containerFileContent)
			imageReference = getDockerfileTargetImage(containerFileContent, imageReference)
			log.Info().Interface("platforms", platforms).Str("image", imageReference).Msg("building container image")

			// build each image and add to manifest
			for _, platform := range platforms {
				containerArchiveFile := filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), platform.Platform("_")+".tar")

				// build
				var buildArgs []string
				buildArgs = append(buildArgs, `buildah-script`)
				buildArgs = append(buildArgs, containerFile)
				ctx.Env["TARGETIMAGE"] = "oci-archive:" + containerArchiveFile
				ctx.Env["TARGETPLATFORM"] = platform.Platform("/")
				ctx.Env["TARGETOS"] = platform.OS
				ctx.Env["TARGETARCH"] = platform.Arch
				ctx.Env["TARGETVARIANT"] = platform.Variant
				command.RunCommand(strings.Join(buildArgs, " "), ctx.Env, ctx.ProjectDir)
			}
		}
	}

	// store image ref
	_ = filesystem.SaveFileText(filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), "image.txt"), imageReference)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildahPackageActionStruct{})
}
