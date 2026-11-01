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
	Cache  []ContainerCache
}

type DiscoverContainerOptions struct {
	Packages []ContainerPackage
}

var DefaultDiscoverContainerOptions = DiscoverContainerOptions{
	Packages: []ContainerPackage{},
}

func DiscoverContainerCandidates(opts *DiscoverContainerOptions) []Executable {
	var result []Executable
	if opts == nil {
		opts = &DefaultDiscoverContainerOptions
	}

	for _, containerImage := range opts.Packages {
		tags, err := registry.FindTagsWithDigest(containerImage.Image)
		if err != nil {
			log.Error().Err(err).Msgf("failed to find tags for image %s", containerImage.Image)
			continue
		}

		cacheMounts := make([]ContainerCache, 0)
		for _, cache := range containerImage.Cache {
			cacheMounts = append(cacheMounts, cache)
		}

		log.Info().Msgf("Discovering container candidates for image %s", containerImage.Image)
		for _, tag := range tags {
			if strings.HasPrefix(tag.Tag, "sha256-") {
				continue
			}
			// FindTagsWithDigest already drops tags without a resolvable digest, keep the guard so a
			// regression there can never produce an unpinned candidate
			if tag.Digest == "" {
				continue
			}

			log.Info().Msgf("Found tag %s for image %s", tag.Tag, containerImage.Image)
			for _, bin := range containerImage.Binary {
				semverVersion := convertToSemver(tag.Tag)

				result = append(result, ContainerCandidate{
					BaseCandidate: BaseCandidate{
						Name:    bin,
						Version: semverVersion,
						Type:    ExecutionContainer,
					},
					Image:      fmt.Sprintf("%s:%s", containerImage.Image, tag.Tag),
					Digest:     tag.Digest,
					ImageCache: cacheMounts,
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
