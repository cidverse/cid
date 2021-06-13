package config

import (
	_ "embed"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

//go:embed cid-main.yaml
var embeddedConfig string

func LoadConfigurationFile(config interface{}, file string) (err error) {
	cfgErr := configor.New(&configor.Config{ENVPrefix: "CID", Silent: true}).Load(config, file)

	if cfgErr != nil {
		log.Warn().Str("file", file).Msg("failed to load configuration > " + cfgErr.Error())
	}

	return cfgErr
}

var Config = struct {
	Paths           PathConfig
	Mode            ExecutionModeType `default:"PREFER_LOCAL"`
	Conventions     ProjectConventions
	Env             map[string]string
	Stages          []WorkflowStage             `yaml:"stages"`
	Actions         map[string][]WorkflowAction `yaml:"actions"`
	Dependencies    map[string]string
	Tools           []ToolExecutableDiscovery `yaml:"tools"`
	ContainerImages []ToolContainerDiscovery  `yaml:"container-images"`
}{}

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact string `default:"dist"`
	Temp     string `default:"tmp"`
	Cache    string `default:""`
}

func LoadConfig(projectDirectory string) {
	// parent config
	err := yaml.Unmarshal([]byte(embeddedConfig), &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load embedded configuration")
	}

	// load
	loadConfigErr := LoadConfigurationFile(&Config, projectDirectory+"/cid.yml")
	if loadConfigErr != nil {
		log.Fatal().Err(loadConfigErr).Msg("failed to parse config")
	}

	if Config.Dependencies == nil {
		Config.Dependencies = make(map[string]string)
	}
}
