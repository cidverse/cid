package changelogcommon

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
)

func depCommit(scope, subject string) *actionsdk.VCSCommit {
	return &actionsdk.VCSCommit{Context: map[string]string{"type": "chore", "scope": scope, "subject": subject}}
}

func subjects(commits []*actionsdk.VCSCommit) []string {
	var out []string
	for _, c := range commits {
		out = append(out, c.Context["subject"])
	}
	return out
}

func TestMergeDependencyUpdatesRenovate(t *testing.T) {
	commits := []*actionsdk.VCSCommit{
		depCommit("deps", "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.40.2 (#506)"),
		depCommit("deps", "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.42.0 (#508)"),
		depCommit("deps", "update module github.com/other/lib to v1.2.0 (#510)"),
		depCommit("deps", "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.49.0 (#517)"),
		depCommit("deps", "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.9.0 (#500)"),
		depCommit("", "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.50.0 (#520)"), // different scope
		depCommit("core", "resolves an issue"),
	}

	result := MergeDependencyUpdates(commits)

	assert.Equal(t, []string{
		"update module gitlab.com/gitlab-org/api/client-go/v3 to v3.49.0 (#500, #506, #508, #517)",
		"update module github.com/other/lib to v1.2.0 (#510)",
		"update module gitlab.com/gitlab-org/api/client-go/v3 to v3.50.0 (#520)",
		"resolves an issue",
	}, subjects(result))
	assert.Equal(t, "4", result[0].Context["merged_count"])
	// original commits are not mutated
	assert.Equal(t, "update module gitlab.com/gitlab-org/api/client-go/v3 to v3.49.0 (#517)", commits[3].Context["subject"])
}

func TestMergeDependencyUpdatesDependabot(t *testing.T) {
	commits := []*actionsdk.VCSCommit{
		depCommit("deps", "bump golang.org/x/net from 0.2.0 to 0.3.0 (#12)"),
		depCommit("deps", "bump golang.org/x/net from 0.1.0 to 0.2.0 (#10)"),
		depCommit("deps", "bump golang.org/x/net from 0.1.0 to 0.2.0 in /tools (#11)"),
	}

	assert.Equal(t, []string{
		"bump golang.org/x/net from 0.1.0 to 0.3.0 (#10, #12)",
		"bump golang.org/x/net from 0.1.0 to 0.2.0 in /tools (#11)",
	}, subjects(MergeDependencyUpdates(commits)))
}

func TestMergeDependencyUpdatesDigest(t *testing.T) {
	commits := []*actionsdk.VCSCommit{
		depCommit("deps", "update actions/checkout digest to abc1234 (#1)"),
		depCommit("deps", "update actions/checkout digest to def5678 (#2)"),
	}

	assert.Equal(t, []string{
		"update actions/checkout digest to def5678 (#1, #2)",
	}, subjects(MergeDependencyUpdates(commits)))
}
