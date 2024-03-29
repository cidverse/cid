package catalog

import (
	"strings"

	"github.com/cidverse/cid/pkg/core/registry"
	"github.com/rs/zerolog/log"
)

func ProcessCatalog(catalog *Config) *Config {
	result := Config{}

	// actions
	for _, sourceAction := range catalog.Actions { //nolint:gocritic
		sourceAction.Repository = ""
		result.Actions = append(result.Actions, sourceAction)
	}

	// images
	for _, sourceImage := range catalog.ContainerImages { //nolint:gocritic
		sourceImage.Repository = ""
		if len(sourceImage.Source.RegistryURL) > 0 {
			tags, err := registry.FindTags(sourceImage.Source.RegistryURL)
			if err != nil {
				log.Fatal().Err(err).Str("repository", sourceImage.Source.RegistryURL).Msg("failed to query tags for repository")
			}

			for _, tag := range tags {
				if strings.HasPrefix(tag.Tag, "sha256-") {
					continue
				}

				version := tagToVersion(tag.Tag)
				image := sourceImage

				// overwrite image
				image.Image = strings.ReplaceAll(image.Image, "${{TAG}}", tag.Tag)

				// find digest
				/*
					// TODO: make this toggleable
					digest, err := registry.GetArtifactDigest(image.Image)
					if err != nil {
						log.Warn().Err(err).Str("repository", sourceImage.Source.RegistryURL).Str("image", image.Image).Msg("failed to query digest for image")
					}
					image.Digest = digest
				*/

				var providedBinary []ProvidedBinary
				for _, p := range image.Provides {
					providedBinary = append(providedBinary, ProvidedBinary{
						Binary:  p.Binary,
						Version: strings.ReplaceAll(p.Version, "${{VERSION}}", version),
						Alias:   p.Alias,
					})
				}
				image.Provides = providedBinary
				image.Certs = sourceImage.Certs
				result.ContainerImages = append(result.ContainerImages, image)
			}
		} else {
			result.ContainerImages = append(result.ContainerImages, sourceImage)
		}
	}

	// workflows
	for _, sourceWorkflow := range catalog.Workflows { //nolint:gocritic
		sourceWorkflow.Repository = ""
		result.Workflows = append(result.Workflows, sourceWorkflow)
	}

	return &result
}

func tagToVersion(input string) string {
	return input
}
