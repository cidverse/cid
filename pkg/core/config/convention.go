package config

type ProjectConventions struct {
	Branching        BranchingConventionType `default:"GitFlow"`
	Commit           CommitConventionType    `default:"ConventionalCommits"`
	PreReleaseSuffix string                  `default:"-rc.{NCI_LASTRELEASE_COMMIT_AFTER_COUNT}"`
}

type BranchingConventionType string

const (
	BranchingGitFlow BranchingConventionType = "GitFlow"
)

type CommitConventionType string

const (
	ConventionalCommits CommitConventionType = "ConventionalCommits"
)
