package config

import (
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
)

func LoadConfigurationFile(config interface{}, file string) (err error) {
	cfgErr := configor.New(&configor.Config{ENVPrefix: "CID", Silent: true}).Load(config, file)

	if cfgErr != nil {
		log.Warn().Str("file",file).Msg("failed to load configuration > " + cfgErr.Error())
	}

	return cfgErr
}

var Config = struct {
	Paths PathConfig
	Conventions ProjectConventions
	Workflow []WorkflowStage
	Mode ExecutionModeType `default:"PREFER_LOCAL"`
	Dependencies map[string]string
}{}

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact string `default:"dist"`
	Cache string `default:""`
}

type ProjectConventions struct {
	Branching BranchingConventionType `default:"GitFlow"`
	Commit CommitConventionType `default:"ConventionalCommits"`
	PreReleaseSuffix string `default:"-rc.{NCI_LASTRELEASE_COMMIT_AFTER_COUNT}"`
}

type WorkflowStage struct {
	Stage string
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{}
}

// ExecutionModeType
type ExecutionModeType string
const(
	PreferLocal ExecutionModeType = "PREFER_LOCAL"
	Strict                        = "STRICT"
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

func LoadConfig(projectDirectory string) {
	// load
	LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
	if Config.Dependencies == nil {
		Config.Dependencies = make(map[string]string)
	}
}
