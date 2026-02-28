package changelogcommon

import (
	"encoding/json"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func TestProjectData() *actionsdk.ProjectExecutionContextV1Response {
	cfg := Config{
		Templates: []string{
			"github.changelog",
		},
		CommitPattern: []string{
			"(?P<type>[A-Za-z]+)((?:\\((?P<scope>[^()\\r\\n]*)\\)|\\()?(?P<breaking>!)?)(:\\s?(?P<subject>.*))?",
		},
		TitleMaps: map[string]string{
			"build":    "Build System",
			"ci":       "CI",
			"docs":     "Documentation",
			"feat":     "Features",
			"fix":      "Bug Fixes",
			"perf":     "Performance",
			"refactor": "Refactor",
			"style":    "Style",
			"test":     "Test",
			"chore":    "Internal",
		},
		NoteKeywords: []NoteKeyword{
			{
				Keyword: "NOTE",
				Title:   "Notes",
			},
			{
				Keyword: "BREAKING CHANGE",
				Title:   "Breaking Changes",
			},
		},
		IssuePrefix: "",
	}
	cfgJson, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	return &actionsdk.ProjectExecutionContextV1Response{
		ProjectDir: "/my-project",
		Config: &actionsdk.ConfigV1Response{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
			Config:      string(cfgJson),
		},
		Modules: nil,
		Env: map[string]string{
			"NCI_REPOSITORY_KIND":        "git",
			"NCI_REPOSITORY_REMOTE":      "https://github.com/cidverse/normalizeci.git",
			"NCI_REPOSITORY_URL":         "https://github.com/cidverse/normalizeci",
			"NCI_REPOSITORY_HOST_SERVER": "github.com",
			"NCI_COMMIT_REF_NAME":        "v1.2.0",
			"NCI_COMMIT_HASH":            "abcdef123456",
			"NCI_COMMIT_REF_VCS":         "refs/tags/v1.2.0",
			"NCI_PROJECT_ID":             "123456",
			"NCI_PROJECT_URL":            "https://github.com/cidverse/normalizeci",
		},
	}
}
