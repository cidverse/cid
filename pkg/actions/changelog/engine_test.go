package changelog

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/commitanalyser"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var config = Config{
	TitleMaps: map[string]string{
		"feat": "Features",
		"fix":  "Bug Fixes",
	},
	CommitPattern: []string{commitanalyser.ConventionalCommitPattern},
	NoteKeywords:  []NoteKeyword{{"NOTE", "Notes"}, {"BREAKING CHANGE", "Breaking Changes"}},
}
var commits = []vcsrepository.Commit{
	{
		Hash:        "",
		Message:     "feat: adds new feature",
		Description: "NOTE: this feature is pretty useful",
		Author:      vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
	{
		Hash:        "",
		Message:     "fix(core): resolves a issue",
		Description: "",
		Author:      vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
	{
		Hash:        "",
		Message:     "fix(core): resolves a different issue",
		Description: "",
		Author:      vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Committer:   vcsrepository.CommitAuthor{Name: "Philipp Heuer", Email: "contact@example.com"},
		Tags:        nil,
	},
}

func TestContributorData(t *testing.T) {
	// preprocess
	commits = PreprocessCommits(config, commits)

	// analyse / grouping
	templateData := ProcessCommits(config, commits)

	assert.Equal(t, "Philipp Heuer", templateData.Contributors["contact@example.com"].Name)
	assert.Equal(t, "contact@example.com", templateData.Contributors["contact@example.com"].Email)
	assert.Equal(t, 3, templateData.Contributors["contact@example.com"].Commits)
}

func TestRenderGitHubReleaseTemplate(t *testing.T) {
	// get template
	template, templateErr := api.GetFileContentFromEmbedFS(TemplateFS, "templates/github-release.tmpl")
	assert.NoError(t, templateErr)

	// preprocess
	commits = PreprocessCommits(config, commits)

	// analyse / grouping
	templateData := ProcessCommits(config, commits)
	templateData.ProjectUrl = "https://github.com/cidverse/cid"
	templateData.ProjectName = "CID"
	templateData.Version = "1.0.0"
	templateData.ReleaseDate = time.Unix(int64(1609502400), int64(0))

	output, outputErr := RenderTemplate(templateData, template)
	assert.NoError(t, outputErr)
	assert.Equal(t, `## Bug Fixes
- **core:** resolves a issue
- **core:** resolves a different issue

## Features
- adds new feature

## Notes
-  this feature is pretty useful
`, output)
}

func TestRenderDiscordTemplate(t *testing.T) {
	// get template
	template, templateErr := api.GetFileContentFromEmbedFS(TemplateFS, "templates/discord-release.tmpl")
	assert.NoError(t, templateErr)

	// preprocess
	commits = PreprocessCommits(config, commits)

	// analyse / grouping
	templateData := ProcessCommits(config, commits)
	templateData.ProjectUrl = "https://github.com/cidverse/cid"
	templateData.ProjectName = "CID"
	templateData.Version = "1.0.0"
	templateData.ReleaseDate = time.Unix(int64(1609502400), int64(0))

	output, outputErr := RenderTemplate(templateData, template)
	assert.NoError(t, outputErr)
	assert.Equal(t, `:rocket: CID - ***1.0.0*** - 2021-01-01 :rocket:

**Bug Fixes**
- **core:** resolves a issue
- **core:** resolves a different issue

**Features**
- adds new feature

**Notes**
-  this feature is pretty useful
`, output)
}
