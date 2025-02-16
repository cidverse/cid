package catalog

import (
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/rs/zerolog/log"
)

func ProcessCatalog(catalog *Config) *Config {
	result := Config{
		Actions:     nil,
		Workflows:   nil,
		Executables: nil,
	}

	// actions
	for _, sourceAction := range catalog.Actions {
		result.Actions = append(result.Actions, sourceAction)
	}

	// workflows
	for _, sourceWorkflow := range catalog.Workflows {
		result.Workflows = append(result.Workflows, sourceWorkflow)
	}

	// executable-discovery
	containerDiscovery := catalog.ExecutableDiscovery.ContainerDiscovery
	if len(containerDiscovery.Packages) > 0 {
		log.Debug().Int("packages", len(containerDiscovery.Packages)).Msg("container packages defined in executable-discovery")
		containerCandidates := executable.DiscoverContainerCandidates(&containerDiscovery)
		for _, cc := range containerCandidates {
			typedCandidate, err := executable.ToTypedCandidate(cc)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to convert container candidate to typed candidate")
			}
			result.Executables = append(result.Executables, typedCandidate)
		}
	} else {
		log.Debug().Msg("no container packages defined in executable-discovery")
	}

	// executables
	for _, sourceExecutable := range catalog.Executables {
		result.Executables = append(result.Executables, sourceExecutable)
	}

	return &result
}
