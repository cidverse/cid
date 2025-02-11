package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/cidverse/cid/pkg/common/candidate"
	"github.com/cidverse/cid/pkg/core/config"
)

var (
	ErrNoCandidatesProvided = fmt.Errorf("candidates are a required field")
	ErrNoCommandProvided    = fmt.Errorf("command is a required field")
)

func CandidatesFromConfig(cfg config.CIDConfig) ([]candidate.Candidate, error) {
	var result []candidate.Candidate

	// path candidates
	var pathDiscoveryRules []candidate.PathDiscoveryRule
	for _, entry := range cfg.LocalTools {
		var pathDiscoveryRulesLookup []candidate.PathDiscoveryRuleLookup

		for _, lookup := range entry.Lookup {
			pathDiscoveryRulesLookup = append(pathDiscoveryRulesLookup, candidate.PathDiscoveryRuleLookup{
				Key:            lookup.Key,
				KeyAliases:     []string{},
				Directory:      "",
				Version:        lookup.Version,
				VersionCommand: "",
				VersionRegex:   "",
			})
		}

		pathDiscoveryRules = append(pathDiscoveryRules, candidate.PathDiscoveryRule{
			Binary: entry.Binary,
			Lookup: pathDiscoveryRulesLookup,
		})
	}
	result = append(result, candidate.DiscoverPathCandidates(&candidate.DiscoverPathOptions{
		LookupRules:          pathDiscoveryRules,
		VersionLookupCommand: false,
	})...)

	// image registry
	for _, entry := range cfg.Registry.ContainerImages {
		for _, provided := range entry.Provides {
			var containerCache []candidate.ContainerCache
			for _, cache := range entry.Cache {
				containerCache = append(containerCache, candidate.ContainerCache{
					ID:            cache.ID,
					ContainerPath: cache.ContainerPath,
					MountType:     cache.MountType,
				})
			}

			var containerMounts []candidate.ContainerMount
			for _, mount := range entry.Mounts {
				containerMounts = append(containerMounts, candidate.ContainerMount{
					Src:  mount.Src,
					Dest: mount.Dest,
				})
			}

			var containerCerts []candidate.ContainerCerts
			for _, cert := range entry.Certs {
				containerCerts = append(containerCerts, candidate.ContainerCerts{
					Type:          cert.Type,
					ContainerPath: cert.ContainerPath,
				})
			}

			result = append(result, candidate.ContainerCandidate{
				BaseCandidate: candidate.BaseCandidate{
					Name:    provided.Binary,
					Version: provided.Version,
					Type:    candidate.ExecutionContainer,
				},
				Image:      entry.Image,
				ImageCache: containerCache,
				Mounts:     containerMounts,
				Security: candidate.ContainerSecurity{
					Capabilities: entry.Security.Capabilities,
					Privileged:   entry.Security.Privileged,
				},
				Entrypoint: entry.Entrypoint,
				Certs:      containerCerts,
			})
		}
	}

	// nix candidates
	result = append(result, candidate.DiscoverNixStoreCandidates(nil)...)

	return result, nil
}

type Opts struct {
	Candidates             []candidate.Candidate
	CandidateTypes         []candidate.CandidateType
	Command                string
	Env                    map[string]string
	ProjectDir             string
	WorkDir                string
	TempDir                string
	CaptureOutput          bool
	Ports                  []int
	UserProvidedConstraint string
	Constraints            map[string]string
	Stdin                  io.Reader
}

// Execute gets called from actions or the api to execute commands
func Execute(opts Opts) (stdout string, stderr string, cand candidate.Candidate, err error) {
	// validate
	if len(opts.Candidates) == 0 {
		return "", "", cand, ErrNoCandidatesProvided
	}
	if len(opts.Command) == 0 {
		return "", "", cand, ErrNoCommandProvided
	}

	// identify command
	args := strings.SplitN(opts.Command, " ", 2)
	cmdBinary := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = strings.Split(args[1], " ")
	}

	// constraint from config
	versionConstraint := candidate.AnyVersionConstraint
	if value, ok := opts.Constraints[cmdBinary]; ok {
		versionConstraint = value
	}
	// user provided constraint
	if len(opts.UserProvidedConstraint) > 0 {
		versionConstraint = opts.UserProvidedConstraint
	}

	// select candidate
	c := candidate.SelectCandidate(opts.Candidates, candidate.CandidateFilter{
		Types:             opts.CandidateTypes,
		Executable:        cmdBinary,
		VersionPreference: candidate.PreferHighest,
		VersionConstraint: versionConstraint,
	})
	if c == nil {
		return "", "", nil, fmt.Errorf("no candidate found for %s fulfilling constraint %s", cmdBinary, versionConstraint)
	}
	cand = *c

	// run command
	stdout, stderr, err = cand.Run(candidate.RunParameters{
		Args:          cmdArgs,
		Env:           opts.Env,
		RootDir:       opts.ProjectDir,
		WorkDir:       opts.WorkDir,
		TempDir:       opts.TempDir,
		CaptureOutput: opts.CaptureOutput,
	})
	if err != nil {
		return stdout, stderr, cand, err
	}

	return stdout, stderr, cand, nil
}
