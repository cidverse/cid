package executable

import (
	"fmt"
	"strings"

	"github.com/cidverse/cid/pkg/core/registry"
	"github.com/rs/zerolog/log"
)

type ContainerPackage struct {
	Binary []string
	Image  string
}

type DiscoverContainerOptions struct {
	Packages []ContainerPackage
}

var DefaultDiscoverContainerOptions = DiscoverContainerOptions{
	Packages: []ContainerPackage{},
}

func DiscoverContainerCandidates(opts *DiscoverContainerOptions) []Candidate {
	var result []Candidate
	if opts == nil {
		opts = &DefaultDiscoverContainerOptions
	}

	for _, containerImage := range opts.Packages {
		tags, err := registry.FindTags(containerImage.Image)
		if err != nil {
			log.Error().Err(err).Msgf("failed to find tags for image %s", containerImage.Image)
		}

		log.Info().Msgf("Discovering container candidates for image %s", containerImage.Image)
		for _, tag := range tags {
			if strings.HasPrefix(tag.Tag, "sha256-") {
				continue
			}

			log.Info().Msgf("Found tag %s for image %s", tag.Tag, containerImage.Image)
			for _, bin := range containerImage.Binary {
				result = append(result, ContainerCandidate{
					BaseCandidate: BaseCandidate{
						Name:    bin,
						Version: tag.Tag,
						Type:    ExecutionContainer,
					},
					Image:      fmt.Sprintf("%s:%s", containerImage.Image, tag.Tag),
					ImageCache: make([]ContainerCache, 0),
					Mounts:     make([]ContainerMount, 0),
					Security:   ContainerSecurity{},
					Entrypoint: nil,
					Certs:      make([]ContainerCerts, 0),
				})
			}
		}
	}

	return result
}
