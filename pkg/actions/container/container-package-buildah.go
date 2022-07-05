package container

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

type BuildahPackageActionStruct struct{}

type BuildahConfig struct {
	NoCache   bool `yaml:"no-cache" required:"true" default:"true"`
	Squash    bool `yaml:"squash" required:"true" default:"true"`
	Rebuild   bool `yaml:"rebuild" required:"true" default:"true"`
	Platforms []Platform
}

// GetDetails retrieves information about the action
func (action BuildahPackageActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "container-package-buildah",
		Version:   "0.1.0",
		UsedTools: []string{"buildah"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildahPackageActionStruct) Check(ctx api.ActionExecutionContext) bool {
	var missingRequirements []api.MissingRequirement

	if ctx.CurrentModule != nil {
		if !funk.Contains(ctx.CurrentModule.Language, analyzerapi.LanguageDockerfile) && !funk.Contains(ctx.CurrentModule.Language, analyzerapi.LanguageBuildahScript) {
			missingRequirements = append(missingRequirements, api.MissingRequirement{Message: "module is not of language " + string(analyzerapi.LanguageDockerfile) + " or " + string(analyzerapi.LanguageDockerfile)})
		}
	} else {
		missingRequirements = append(missingRequirements, api.MissingRequirement{Message: "no module context present"})
	}

	return len(missingRequirements) == 0
}

// Execute runs the action
func (action BuildahPackageActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	var config BuildahConfig
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	containerFile := strings.TrimPrefix(ctx.CurrentModule.Discovery, "file~")
	image := getFullImage(ctx.Env["NCI_CONTAINERREGISTRY_HOST"], ctx.Env["NCI_CONTAINERREGISTRY_REPOSITORY"], ctx.Env["NCI_CONTAINERREGISTRY_TAG"])

	if funk.Contains(ctx.CurrentModule.Language, analyzerapi.LanguageDockerfile) {
		dockerfileContent, _ := filesystem.GetFileContent(containerFile)
		syntax := getDockerfileSyntax(dockerfileContent)
		platforms := getDockerfileTargetPlatforms(dockerfileContent)
		targetImage := getDockerfileTargetImage(dockerfileContent)
		if len(targetImage) > 0 {
			image = targetImage
		}
		log.Info().Str("syntax", syntax).Interface("platforms", platforms).Str("image", image).Msg("building container image")

		// skip image generation, if image is present in remote registry
		if config.Rebuild != true {
			_, remoteImageErr := LoadRemoteImageInformation(image)
			if remoteImageErr == nil {
				log.Info().Str("syntax", syntax).Interface("platforms", platforms).Str("image", image).Str("cause", "present_in_remote").Msg("skipping container image build")
				return nil
			}
		}

		// old manifest needs to be deleted first if it was present
		command.RunCommand(`buildah manifest rm `+image+` > /dev/null 2>&1 || return 0`, ctx.Env, ctx.ProjectDir)

		// build each image and add to manifest
		for _, platform := range platforms {
			var buildArgs []string
			buildArgs = append(buildArgs, `buildah bud`)
			buildArgs = append(buildArgs, `--os `+platform.OS)
			buildArgs = append(buildArgs, `--arch `+platform.Arch)

			// options
			if config.NoCache {
				buildArgs = append(buildArgs, `--no-cache`)
			}
			if config.Squash {
				buildArgs = append(buildArgs, `--squash`) // squash, excluding the base layer
			}

			// manifest creation for multi-platform images
			if len(platforms) > 1 {
				buildArgs = append(buildArgs, `--manifest `+image)
			} else {
				buildArgs = append(buildArgs, `-t `+image)
			}

			// labels
			buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.source=`+strings.TrimSuffix(ctx.Env["NCI_REPOSITORY_REMOTE"], ".git")+`"`)
			buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.created=`+time.Now().Format(time.RFC3339)+`"`)
			buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.authors="`)
			buildArgs = append(buildArgs, `--annotation "org.opencontainers.image.title=`+ctx.CurrentModule.Name+`"`)

			// dynamic build-args
			if strings.Contains(dockerfileContent, "ARG TARGETPLATFORM") {
				buildArgs = append(buildArgs, `--build-arg TARGETPLATFORM=`+platform.OS+`/`+platform.OS)
			}
			if strings.Contains(dockerfileContent, "ARG TARGETOS") {
				buildArgs = append(buildArgs, `--build-arg TARGETOS=`+platform.OS)
			}
			if strings.Contains(dockerfileContent, "ARG TARGETARCH") {
				buildArgs = append(buildArgs, `--build-arg TARGETARCH=`+platform.Arch)
			}

			buildArgs = append(buildArgs, ctx.CurrentModule.Directory)

			command.RunCommand(strings.Join(buildArgs, " "), ctx.Env, ctx.ProjectDir)
		}

		// push image (sign image if possible)
		var pushArgs []string
		pushArgs = append(pushArgs, `buildah push`)

		// pushArgs = append(pushArgs, `--sign-by test`)
		if len(platforms) > 1 {
			pushArgs = append(pushArgs, `--all`)
		}
		pushArgs = append(pushArgs, image)
		command.RunCommand(strings.Join(pushArgs, " "), ctx.Env, ctx.ProjectDir)
	} else if funk.Contains(ctx.CurrentModule.Language, analyzerapi.LanguageBuildahScript) {
		log.Info().Str("image", image).Str("script", containerFile).Msg("building container image")
	}

	// publish image
	if len(ctx.Env["NCI_CONTAINERREGISTRY_HOST"]) > 0 {
		// command.RunCommand("docker push "+image, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildahPackageActionStruct{})
}