package candidate

import (
	"strings"
)

type ContainerPackage struct {
	Binary []string
	Image  string
}

type DiscoverContainerOptions struct {
	Packages []ContainerPackage
}

var DefaultDiscoverContainerOptions = DiscoverContainerOptions{
	Packages: []ContainerPackage{
		{
			Binary: []string{"syft"},
			Image:  "ghcr.io/cidverse/syft:1.18.1",
		},
	},
}

func DiscoverContainerCandidates(opts *DiscoverContainerOptions) []Candidate {
	var result []Candidate
	if opts == nil {
		opts = &DefaultDiscoverContainerOptions
	}

	// discover using store paths
	for _, containerImage := range opts.Packages {
		version := strings.Split(containerImage.Image, ":")[1]

		for _, bin := range containerImage.Binary {
			result = append(result, ContainerCandidate{
				BaseCandidate: BaseCandidate{
					Name:    bin,
					Version: version,
					Type:    ExecutionContainer,
				},
				Image:      containerImage.Image,
				ImageCache: make([]ContainerCache, 0),
				Mounts:     make([]ContainerMount, 0),
				Security:   ContainerSecurity{},
				Entrypoint: nil,
				Certs:      make([]ContainerCerts, 0),
			})
		}
	}

	return result
}
