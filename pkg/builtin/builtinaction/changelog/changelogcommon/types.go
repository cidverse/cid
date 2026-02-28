package changelogcommon

import (
	"embed"
	"time"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

//go:embed templates/*
var TemplateFS embed.FS

type TemplateData struct {
	ProjectName  string
	ProjectURL   string
	Version      string
	ReleaseDate  time.Time
	Commits      []*actionsdk.VCSCommit
	CommitGroups map[string][]*actionsdk.VCSCommit
	NoteGroups   map[string][]string
	Contributors map[string]ContributorData
}

type ContributorData struct {
	Name    string
	Email   string
	Commits int
}

type Config struct {
	Templates     []string          `yaml:"templates"`
	CommitPattern []string          `yaml:"commit_pattern"`
	TitleMaps     map[string]string `yaml:"title_maps"`
	NoteKeywords  []NoteKeyword     `yaml:"note_keywords"`
	IssuePrefix   string            `yaml:"issue_prefix"`
}

type NoteKeyword struct {
	Keyword string
	Title   string
}
