package config

type ProjectConventions struct {
	Branching BranchingConventionType `default:"GitFlow"`
	Commit CommitConventionType `default:"ConventionalCommits"`
	PreReleaseSuffix string `default:"-rc.{NCI_LASTRELEASE_COMMIT_AFTER_COUNT}"`
}

// ExecutionModeType
type ExecutionModeType string
const(
	PreferLocal ExecutionModeType = "PREFER_LOCAL"
	Strict ExecutionModeType      = "STRICT"
)

// BranchingConventionType
type BranchingConventionType string
const(
	BranchingGitFlow BranchingConventionType = "GitFlow"
)

// BranchingConventionType
type CommitConventionType string
const(
	ConventionalCommits CommitConventionType = "ConventionalCommits"
)
