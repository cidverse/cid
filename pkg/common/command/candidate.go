package command

import (
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/config"
)

func CandidatesFromConfig(cfg config.CIDConfig) ([]executable.Candidate, error) {
	// load candidates
	result, err := executable.LoadCachedExecutables()
	if err != nil {
		return nil, err
	}

	// append config candidates - path
	var pathDiscoveryRules []executable.PathDiscoveryRule
	for _, entry := range cfg.LocalTools {
		var pathDiscoveryRulesLookup []executable.PathDiscoveryRuleLookup

		for _, lookup := range entry.Lookup {
			pathDiscoveryRulesLookup = append(pathDiscoveryRulesLookup, executable.PathDiscoveryRuleLookup{
				Key:            lookup.Key,
				KeyAliases:     []string{},
				Directory:      "",
				Version:        lookup.Version,
				VersionCommand: "",
				VersionRegex:   "",
			})
		}

		pathDiscoveryRules = append(pathDiscoveryRules, executable.PathDiscoveryRule{
			Binary: entry.Binary,
			Lookup: pathDiscoveryRulesLookup,
		})
	}
	result = append(result, executable.DiscoverPathCandidates(&executable.DiscoverPathOptions{
		LookupRules:          pathDiscoveryRules,
		VersionLookupCommand: false,
	})...)

	// append registry candidates - image registry
	for _, entry := range cfg.Registry.ContainerImages {
		for _, provided := range entry.Provides {
			var containerCache []executable.ContainerCache
			for _, cache := range entry.Cache {
				containerCache = append(containerCache, executable.ContainerCache{
					ID:            cache.ID,
					ContainerPath: cache.ContainerPath,
					MountType:     cache.MountType,
				})
			}

			var containerMounts []executable.ContainerMount
			for _, mount := range entry.Mounts {
				containerMounts = append(containerMounts, executable.ContainerMount{
					Src:  mount.Src,
					Dest: mount.Dest,
				})
			}

			var containerCerts []executable.ContainerCerts
			for _, cert := range entry.Certs {
				containerCerts = append(containerCerts, executable.ContainerCerts{
					Type:          cert.Type,
					ContainerPath: cert.ContainerPath,
				})
			}

			result = append(result, executable.ContainerCandidate{
				BaseCandidate: executable.BaseCandidate{
					Name:    provided.Binary,
					Version: provided.Version,
					Type:    executable.ExecutionContainer,
				},
				Image:      entry.Image,
				ImageCache: containerCache,
				Mounts:     containerMounts,
				Security: executable.ContainerSecurity{
					Capabilities: entry.Security.Capabilities,
					Privileged:   entry.Security.Privileged,
				},
				Entrypoint: entry.Entrypoint,
				Certs:      containerCerts,
			})
		}
	}

	return result, nil
}
