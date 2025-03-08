package executable

import (
	"io"
	"slices"
	"sort"

	"github.com/cidverse/cidverseutils/version"
)

type CandidateType string

const (
	ExecutionExec      CandidateType = "exec"
	ExecutionContainer CandidateType = "container"
	ExecutionNixStore  CandidateType = "nix-store"
	ExecutionNixShell CandidateType = "nix-shell"
)

type RunParameters struct {
	Executable    string
	Args          []string
	Env           map[string]string
	RootDir       string
	WorkDir       string
	TempDir       string
	Ports         []int
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	CaptureOutput bool
}

type Executable interface {
	GetName() string
	GetVersion() string
	GetType() CandidateType
	GetUri() string // GetUri returns the URI of the candidate, for auditing purposes
	Run(opts RunParameters) (string, string, error)
}

type BaseCandidate struct {
	Name    string        `yaml:"name,required"`
	Version string        `yaml:"version,required"`
	Type    CandidateType `yaml:"type,required"`
}

func (c BaseCandidate) GetName() string {
	return c.Name
}

func (c BaseCandidate) GetVersion() string {
	return c.Version
}

func (c BaseCandidate) GetType() CandidateType {
	return c.Type
}

func (c BaseCandidate) GetUri() string {
	return ""
}

func (c BaseCandidate) Run(opts RunParameters) (string, string, error) {
	return "", "", nil
}

func ToCandidateTypes(types []string) []CandidateType {
	var result []CandidateType
	for _, t := range types {
		result = append(result, CandidateType(t))
	}
	return result
}

type CandidateFilter struct {
	Types             []CandidateType
	Executable        string
	VersionPreference PreferVersion
	VersionConstraint string
}

// SelectCandidate selects the first candidate that matches the given requirements
func SelectCandidate(candidates []Executable, options CandidateFilter) *Executable {
	var filteredCandidates []Executable
	for _, candidate := range candidates {
		// type constraint
		if len(options.Types) > 0 && !slices.Contains(options.Types, candidate.GetType()) {
			continue
		}

		// executable constraint
		if len(options.Executable) > 0 && candidate.GetName() != options.Executable {
			continue
		}

		// version constraint
		if !version.FulfillsConstraint(candidate.GetVersion(), options.VersionConstraint) {
			continue
		}

		filteredCandidates = append(filteredCandidates, candidate)
	}

	// filter by type
	var orderedCandidates []Executable
	if len(options.Types) == 0 {
		orderedCandidates = filteredCandidates
	} else {
		for _, t := range options.Types {
			for _, candidate := range filteredCandidates {
				if candidate.GetType() == t {
					orderedCandidates = append(orderedCandidates, candidate)
				}
			}
		}
	}
	if len(orderedCandidates) == 0 {
		return nil
	}

	// sort
	sort.Slice(orderedCandidates, func(i, j int) bool {
		// sort by version
		if options.VersionPreference == PreferHighest {
			result, _ := version.Compare(orderedCandidates[i].GetVersion(), orderedCandidates[j].GetVersion())
			return result > 0
		} else {
			result, _ := version.Compare(orderedCandidates[i].GetVersion(), orderedCandidates[j].GetVersion())
			return result < 0
		}
	})

	return &orderedCandidates[0]
}
