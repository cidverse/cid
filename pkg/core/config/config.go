package config

import (
	"embed"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

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

var Current = CIDConfig{}

func LoadConfig(projectDirectory string) *CIDConfig {
	cfg := CIDConfig{}
	// internal config
	unmarshalNoError("files/cid-main.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-main.yaml")), &cfg))
	unmarshalNoError("files/cid-tools.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-tools.yaml")), &cfg))
	unmarshalNoError("files/cid-container.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-container.yaml")), &cfg))
	unmarshalNoError("files/cid-catalog-actions.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-catalog-actions.yaml")), &cfg))
	unmarshalNoError("files/cid-workflow-main.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-workflow-main.yaml")), &cfg))

	// load project config
	if filesystem.FileExists(projectDirectory + "/cid.yml") {
		loadConfigErr := LoadConfigurationFile(&cfg, projectDirectory+"/cid.yml")
		if loadConfigErr != nil {
			log.Err(loadConfigErr).Msg("failed to parse config")
		}
	}

	if cfg.Dependencies == nil {
		cfg.Dependencies = make(map[string]string)
	}

	Current = cfg
	return &cfg
}

func unmarshalNoError(file string, err error) {
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("failed to parse internal yaml configuration!")
	}
}

func getEmbeddedConfig(name string) string {
	var content, _ = embeddedConfigFS.ReadFile(name)
	return string(content)
}
