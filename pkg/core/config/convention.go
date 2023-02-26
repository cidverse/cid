package config

type ProjectConventions struct {
	Branching BranchingConventionType `default:"GitFlow"`
	Commit    CommitConventionType    `default:"ConventionalCommits"`
}

type BranchingConventionType string

const (
	BranchingGitFlow BranchingConventionType = "GitFlow"
)

type CommitConventionType string

const (
	ConventionalCommits CommitConventionType = "ConventionalCommits"
)
