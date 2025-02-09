package candidate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectCandidateByTypeFilter(t *testing.T) {
	candidate := SelectCandidate([]Candidate{
		newCandidate("helm", ExecutionExec, "1.0.0"),
		newCandidate("kubectl", ExecutionExec, "1.0.0"),
		newCandidate("kubectl", ExecutionContainer, "1.0.0"),
	}, CandidateFilter{
		Types:             []CandidateType{ExecutionExec},
		Executable:        "kubectl",
		VersionPreference: PreferHighest,
		VersionConstraint: AnyVersionConstraint,
	})
	assert.NotNil(t, candidate)
	assert.Equal(t, "kubectl", (*candidate).GetName())
	assert.Equal(t, ExecutionExec, (*candidate).GetType())
}

func TestSelectCandidateByVersionFilter(t *testing.T) {
	candidate := SelectCandidate([]Candidate{
		newCandidate("helm", ExecutionExec, "1.0.0"),
		newCandidate("kubectl", ExecutionExec, "1.0.0"),
		newCandidate("kubectl", ExecutionContainer, "2.0.0"),
	}, CandidateFilter{
		Types:             []CandidateType{ExecutionExec, ExecutionContainer},
		Executable:        "kubectl",
		VersionPreference: PreferHighest,
		VersionConstraint: ">= 2.0.0",
	})
	assert.NotNil(t, candidate)
	assert.Equal(t, "kubectl", (*candidate).GetName())
	assert.Equal(t, ExecutionContainer, (*candidate).GetType())
}

func newCandidate(name string, candidateType CandidateType, version string) Candidate {
	return &BaseCandidate{
		Name:    name,
		Type:    candidateType,
		Version: version,
	}
}
