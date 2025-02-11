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
	Packages: []ContainerPackage{
		{
			Binary: []string{"kubectl"},
			Image:  "ghcr.io/cidverse/kubectl",
		},
		{
			Binary: []string{"mockery"},
			Image:  "ghcr.io/cidverse/mockery",
		},
		{
			Binary: []string{"twitch"},
			Image:  "ghcr.io/cidverse/twitch-cli",
		},
		{
			Binary: []string{"syft"},
			Image:  "ghcr.io/cidverse/syft",
		},
		{
			Binary: []string{"wrangler"},
			Image:  "ghcr.io/cidverse/wrangler",
		},
		{
			Binary: []string{"shellcheck"},
			Image:  "ghcr.io/cidverse/shellcheck",
		},
		{
			Binary: []string{"flake8"},
			Image:  "ghcr.io/cidverse/flake8",
		},
		{
			Binary: []string{"renovate"},
			Image:  "ghcr.io/cidverse/renovate",
		},
		{
			Binary: []string{"oc"},
			Image:  "ghcr.io/cidverse/openshift",
		},
		{
			Binary: []string{"poetry"},
			Image:  "ghcr.io/cidverse/poetry",
		},
		{
			Binary: []string{"pipenv"},
			Image:  "ghcr.io/cidverse/pipenv",
		},
		{
			Binary: []string{"aws"},
			Image:  "ghcr.io/cidverse/aws",
		},
		{
			Binary: []string{"gosec"},
			Image:  "ghcr.io/cidverse/gosec",
		},
		{
			Binary: []string{"ansible"},
			Image:  "ghcr.io/cidverse/ansible",
		},
		{
			Binary: []string{"ansible-lint"},
			Image:  "ghcr.io/cidverse/ansible-lint",
		},
		{
			Binary: []string{"osv-scanner"},
			Image:  "ghcr.io/cidverse/osv-scanner",
		},
		{
			Binary: []string{"upx"},
			Image:  "ghcr.io/cidverse/upx",
		},
		{
			Binary: []string{"gh"},
			Image:  "ghcr.io/cidverse/gh",
		},
		{
			Binary: []string{"gitleaks"},
			Image:  "ghcr.io/cidverse/gitleaks",
		},
		{
			Binary: []string{"glab"},
			Image:  "ghcr.io/cidverse/glab",
		},
		{
			Binary: []string{"oras"},
			Image:  "ghcr.io/cidverse/oras",
		},
		{
			Binary: []string{"slsa-verifier"},
			Image:  "ghcr.io/cidverse/slsa-verifier",
		},
		{
			Binary: []string{"kubeseal"},
			Image:  "ghcr.io/cidverse/kubeseal",
		},
		{
			Binary: []string{"semgrep"},
			Image:  "ghcr.io/cidverse/semgrep",
		},
		{
			Binary: []string{"scorecard"},
			Image:  "ghcr.io/cidverse/scorecard",
		},
		{
			Binary: []string{"hugo"},
			Image:  "ghcr.io/cidverse/hugo",
		},
		{
			Binary: []string{"grype"},
			Image:  "ghcr.io/cidverse/grype",
		},
		{
			Binary: []string{"ggshield"},
			Image:  "ghcr.io/cidverse/ggshield",
		},
		{
			Binary: []string{"runpodctl"},
			Image:  "ghcr.io/cidverse/runpodctl",
		},
		{
			Binary: []string{"cue"},
			Image:  "ghcr.io/cidverse/cue",
		},
		{
			Binary: []string{"hadolint"},
			Image:  "ghcr.io/cidverse/hadolint",
		},
		{
			Binary: []string{"liquibase"},
			Image:  "ghcr.io/cidverse/liquibase",
		},
		{
			Binary: []string{"cosign"},
			Image:  "ghcr.io/cidverse/cosign",
		},
		{
			Binary: []string{"rekor-cli"},
			Image:  "ghcr.io/cidverse/rekor-cli",
		},
		{
			Binary: []string{"buildah"},
			Image:  "ghcr.io/cidverse/buildah",
		},
		{
			Binary: []string{"helmfile"},
			Image:  "ghcr.io/cidverse/helmfile",
		},
		{
			Binary: []string{"mc"},
			Image:  "ghcr.io/cidverse/minio-client",
		},
		{
			Binary: []string{"rundeck-cli"},
			Image:  "ghcr.io/cidverse/rundeck-cli",
		},
		{
			Binary: []string{"appinspector"},
			Image:  "ghcr.io/cidverse/appinspector",
		},
		{
			Binary: []string{"codecov"},
			Image:  "ghcr.io/cidverse/codecov-cli",
		},
		{
			Binary: []string{"sonar-scanner"},
			Image:  "ghcr.io/cidverse/sonarscanner-cli",
		},
		{
			Binary: []string{"helm"},
			Image:  "ghcr.io/cidverse/helm",
		},
		{
			Binary: []string{"fossa-cli"},
			Image:  "ghcr.io/cidverse/fossa-cli",
		},
		{
			Binary: []string{"sarifrs"},
			Image:  "ghcr.io/cidverse/sarifrs",
		},
		{
			Binary: []string{"skopeo"},
			Image:  "ghcr.io/cidverse/skopeo",
		},
		{
			Binary: []string{"zizmor"},
			Image:  "ghcr.io/cidverse/zizmor",
		},
	},
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

		for _, tag := range tags {
			if strings.HasPrefix(tag.Tag, "sha256-") {
				continue
			}

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
