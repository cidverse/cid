package config

import (
	"embed"
	_ "embed"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

//go:embed files/cid-main.yaml
var embeddedConfig string

//go:embed files
var embeddedConfigFS embed.FS

func LoadConfigurationFile(config interface{}, file string) (err error) {
	log.Debug().Str("file", file).Msg("loading configuration file ...")
	cfgErr := configor.New(&configor.Config{ENVPrefix: "CID", Silent: true}).Load(config, file)

	if cfgErr != nil {
		log.Warn().Str("file", file).Msg("failed to load configuration > " + cfgErr.Error())
	}

	return cfgErr
}

var Config = CIDConfig{}

func LoadConfig(projectDirectory string) {
	// internal config
	yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-main.yaml")), &Config)
	yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-tools.yaml")), &Config)
	yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-container.yaml")), &Config)

	// load project config
	loadConfigErr := LoadConfigurationFile(&Config, projectDirectory+"/cid.yml")
	if loadConfigErr != nil {
		log.Err(loadConfigErr).Msg("failed to parse config")
	}

	if Config.Dependencies == nil {
		Config.Dependencies = make(map[string]string)
	}
}

func getEmbeddedConfig(name string) string {
	var content, _ = embeddedConfigFS.ReadFile(name)
	return string(content)
}
