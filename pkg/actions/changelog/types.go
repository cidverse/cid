package changelog

import (
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"time"
)

type TemplateData struct {
	ProjectName  string
	ProjectUrl   string
	Version      string
	ReleaseDate  time.Time
	Commits      []vcsrepository.Commit
	CommitGroups map[string][]vcsrepository.Commit
	NoteGroups   map[string][]string
	Contributors map[string]ContributorData
}

type ContributorData struct {
	Name    string
	Email   string
	Commits int
}
