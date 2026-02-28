package actionsdk

import (
	"time"

	"github.com/cidverse/go-vcs/vcsapi"
)

type VCSCommit struct {
	HashShort   string            `json:"hash_short,omitempty"`
	Hash        string            `json:"hash,omitempty"`
	Message     string            `json:"message,omitempty"`
	Description string            `json:"description,omitempty"`
	Author      VCSAuthor         `json:"author,omitempty"`
	Committer   VCSAuthor         `json:"committer,omitempty"`
	Tags        []*VCSTag         `json:"tags,omitempty"`
	AuthoredAt  time.Time         `json:"authored_at,omitempty"`
	CommittedAt time.Time         `json:"committed_at,omitempty"`
	Changes     []*VCSChange      `json:"changes,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

type VCSAuthor struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type VCSTag struct {
	RefType string `json:"type,omitempty"`
	Value   string `json:"value,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

type VCSRelease struct {
	Version string        `json:"version,omitempty"`
	Ref     vcsapi.VCSRef `json:"ref,omitempty"`
}

type VCSChange struct {
	ChangeType string  `json:"type,omitempty"`
	FileFrom   VCSFile `json:"file_from,omitempty"`
	FileTo     VCSFile `json:"file_to,omitempty"`
	Patch      string  `json:"patch,omitempty"`
}

type VCSFile struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type VCSDiff struct {
	FileFrom VCSFile       `json:"file_from"`
	FileTo   VCSFile       `json:"file_to"`
	Lines    []VCSDiffLine `json:"lines,omitempty"`
}

type VCSDiffLine struct {
	Operation int    `json:"operation"`
	Content   string `json:"content"`
}

type VCSCommitsRequest struct {
	FromHash       string `json:"from"`
	ToHash         string `json:"to"`
	IncludeChanges bool   `json:"changes"`
	Limit          int    `json:"limit"`
}

type VCSCommitByHashRequest struct {
	Hash           string `json:"hash"`
	IncludeChanges bool   `json:"changes"`
}

type VCSReleasesRequest struct {
	Type string `json:"type"` // Type of the release: stable, unstable
}

type VCSDiffRequest struct {
	FromHash string `json:"from"`
	ToHash   string `json:"to"`
}
