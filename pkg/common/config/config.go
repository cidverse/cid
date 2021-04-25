package config

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
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
	Paths api.PathConfig
	Workflow []WorkflowStage
	Mode ExecutionModeType `default:"PREFER_LOCAL"`
	Dependencies map[string]string
}{}

type WorkflowStage struct {
	Stage string
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{}
}

type ExecutionModeType string
const(
	PreferLocal ExecutionModeType = "PREFER_LOCAL"
	Strict                        = "STRICT"
)

func LoadConfig(projectDirectory string) {
	LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}
