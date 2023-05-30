package commitanalyser

import (
	"testing"

	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/stretchr/testify/assert"
)

func TestDeterminateNextReleaseVersionBreaking(t *testing.T) {
	var commits = []vcsapi.Commit{
		{
			Message:     `fix!: resolves a issue`,
			Description: ``,
		},
	}

	var nextVersion, _ = DeterminateNextReleaseVersion(commits, []string{ConventionalCommitPattern}, DefaultReleaseVersionRules, "v1.0.0")
	assert.Equal(t, "2.0.0", nextVersion, "they should be equal")
}

func TestDeterminateNextReleaseVersionFeature(t *testing.T) {
	var commits = []vcsapi.Commit{
		{
			Message:     `feat: adds new feature`,
			Description: ``,
		},
		{
			Message:     `fix: resolves a issue`,
			Description: ``,
		},
	}

	var nextVersion, _ = DeterminateNextReleaseVersion(commits, []string{ConventionalCommitPattern}, DefaultReleaseVersionRules, "v1.0.0")
	assert.Equal(t, "1.1.0", nextVersion, "they should be equal")
}

func TestDeterminateNextReleaseVersionFix(t *testing.T) {
	var commits = []vcsapi.Commit{
		{
			Message:     `fix: resolves a issue`,
			Description: ``,
		},
	}

	var nextVersion, _ = DeterminateNextReleaseVersion(commits, []string{ConventionalCommitPattern}, DefaultReleaseVersionRules, "v1.0.0")
	assert.Equal(t, "1.0.1", nextVersion, "they should be equal")
}
