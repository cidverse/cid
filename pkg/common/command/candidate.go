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

	return result, nil
}
