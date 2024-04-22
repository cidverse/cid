package config

import (
	"embed"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/filesystem"
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

	// default os cache dir
	catalogSources := catalog.LoadSources()
	cfg.CatalogSources = catalogSources
	data := catalog.LoadCatalogs(catalogSources)
	log.Info().Int("images", len(data.ContainerImages)).Int("actions", len(data.Actions)).Int("workflows", len(data.Workflows)).Msg("loaded catalogs")
	cfg.Registry.ContainerImages = append(cfg.Registry.ContainerImages, data.ContainerImages...)
	cfg.Registry.Actions = append(cfg.Registry.Actions, data.Actions...)
	cfg.Registry.Workflows = append(cfg.Registry.Workflows, data.Workflows...)

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
