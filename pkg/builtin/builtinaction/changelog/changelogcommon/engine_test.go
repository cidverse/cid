package changelogcommon

import (
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/stretchr/testify/assert"
)

var config = Config{
	TitleMaps: map[string]string{
		"feat": "Features",
		"fix":  "Bug Fixes",
	},
	CommitPattern: []string{`(?P<type>[A-Za-z]+)((?:\((?P<scope>[^()\r\n]*)\)|\()?(?P<breaking>!)?)(:\s?(?P<subject>.*))?`},
	NoteKeywords:  []NoteKeyword{{"NOTE", "Notes"}, {"BREAKING CHANGE", "Breaking Changes"}},
}

var commits = []*actionsdk.VCSCommit{
	{
		Hash:        "",
		Message:     "feat: adds new feature",
		Description: "NOTE: this feature is pretty useful",
		Author:      actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
	{
		Hash:        "",
		Message:     "fix(core): resolves a issue",
		Description: "",
		Author:      actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
	{
		Hash:        "",
		Message:     "fix(core): resolves a different issue",
		Description: "",
		Author:      actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   actionsdk.VCSAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
}

func TestContributorData(t *testing.T) {
	// preprocess
	commits = PreprocessCommits(config.CommitPattern, commits)

	// analyze / grouping
	templateData := ProcessCommits(config, commits)

	assert.Equal(t, "Philipp Heuer", templateData.Contributors["contact@example.com"].Name)
	assert.Equal(t, "contact@example.com", templateData.Contributors["contact@example.com"].Email)
	assert.Equal(t, 3, templateData.Contributors["contact@example.com"].Commits)
}

func TestRenderGitHubReleaseTemplate(t *testing.T) {
	// get template
	template, templateErr := GetFileContentFromEmbedFS(TemplateFS, "templates/github.changelog")
	assert.NoError(t, templateErr)

	// preprocess
	commits = PreprocessCommits(config.CommitPattern, commits)

	// analyze / grouping
	templateData := ProcessCommits(config, commits)
	templateData.ProjectURL = "https://github.com/cidverse/cid"
	templateData.ProjectName = "CID"
	templateData.Version = "1.0.0"
	templateData.ReleaseDate = time.Unix(int64(1609502400), int64(0))

	output, outputErr := RenderTemplate(&templateData, template)
	assert.NoError(t, outputErr)
	assert.Equal(t, `## Bug Fixes
- **core:** resolves a different issue
- **core:** resolves a issue

## Features
- adds new feature

## Notes
-  this feature is pretty useful
`, output)
}

func TestRenderDiscordTemplate(t *testing.T) {
	// get template
	template, templateErr := GetFileContentFromEmbedFS(TemplateFS, "templates/discord.changelog")
	assert.NoError(t, templateErr)

	// preprocess
	commits = PreprocessCommits(config.CommitPattern, commits)

	// analyze / grouping
	templateData := ProcessCommits(config, commits)
	templateData.ProjectURL = "https://github.com/cidverse/cid"
	templateData.ProjectName = "CID"
	templateData.Version = "1.0.0"
	templateData.ReleaseDate = time.Unix(int64(1609502400), int64(0))

	output, outputErr := RenderTemplate(&templateData, template)
	assert.NoError(t, outputErr)
	assert.Equal(t, `:rocket: CID - ***1.0.0*** - 2021-01-01 :rocket:

**Bug Fixes**
- **core:** resolves a different issue
- **core:** resolves a issue

**Features**
- adds new feature

**Notes**
-  this feature is pretty useful
`, output)
}
