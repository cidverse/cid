package commitanalyser

// ConventionalCommitPattern is a regex pattern of the Conventional Commits spec - https://www.conventionalcommits.org/en/v1.0.0/
var ConventionalCommitPattern = `(?P<type>[A-Za-z]+)((?:\((?P<scope>[^()\r\n]*)\)|\()?(?P<breaking>!)?)(:\s?(?P<subject>.*))?`

type CommitVersionRule struct {
	Type    string
	Scope   string
	Release string // major / minor / patch
}

var DefaultReleaseVersionRules = []CommitVersionRule{
	{
		Type:    `feat`,
		Release: `minor`,
	},
	{
		Type:    `refactor`,
		Release: `minor`,
	},
	{
		Type:    `fix`,
		Release: `patch`,
	},
	{
		Type:    `ci`,
		Release: `patch`,
	},
	{
		Type:    `build`,
		Release: `patch`,
	},
	{
		Type:    `docs`,
		Release: `patch`,
	},
	{
		Type:    `perf`,
		Release: `patch`,
	},
	{
		Type:    `test`,
		Release: `patch`,
	},
	{
		Type:    `style`,
		Release: `patch`,
	},
	{
		Type:    `chore`,
		Release: `patch`,
	},
}
