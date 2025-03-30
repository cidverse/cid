package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCliffChangelog(t *testing.T) {
	json := `[
	  {
		"version": "v0.0.22",
		"commits": [
		  {
			"id": "561daee32669332fa538cd55793cd24dcdbff3f9",
			"message": "add api.scorecard.dev:443 to scan whitelist",
			"group": "feat"
		  }
		],
		"commit_id": "561daee32669332fa538cd55793cd24dcdbff3f9"
	  }
	]`
	expected := []ChangelogVersion{
		{
			Version: "0.0.22",
			Commits: []Commit{
				{
					ID:      "561daee32669332fa538cd55793cd24dcdbff3f9",
					Message: "add api.scorecard.dev:443 to scan whitelist",
					Group:   "feat",
				},
			},
			CommitID: "561daee32669332fa538cd55793cd24dcdbff3f9",
		},
	}

	result, err := ParseCliffChangelog(json)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
