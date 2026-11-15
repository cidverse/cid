package config

import (
	"embed"

	"github.com/cidverse/cid/pkg/builtin/builtincatalog"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v4"
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

// Current is the configuration of the last LoadConfig call, used by the CLI runtime (e.g. action executors).
// It is process-global: code that loads configs for multiple projects concurrently (cid app-server) must use Load instead.
var Current = CIDConfig{}

// LoadConfig loads the configuration of the project and stores it in Current.
func LoadConfig(projectDirectory string) *CIDConfig {
	cfg := Load(projectDirectory)
	Current = *cfg
	return cfg
}

// Load loads the configuration of the project (embedded defaults, catalogs, cid.yml) without modifying Current.
func Load(projectDirectory string) *CIDConfig {
	cfg := CIDConfig{}

	// defaults
	unmarshalNoError("files/cid-main.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-main.yaml")), &cfg))
	unmarshalNoError("files/cid-tools.yaml", yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-tools.yaml")), &cfg))

	// default os cache dir
	catalogSources := catalog.LoadSources()
	cfg.CatalogSources = catalogSources
	data := catalog.LoadCatalogs(catalogSources)
	log.Debug().Int("catalogs", len(cfg.CatalogSources)).Int("action", len(data.Actions)).Int("workflows", len(data.Workflows)).Int("executables", len(data.Executables)).Msg("loaded catalog from registries")

	cfg.Registry.Actions = append(cfg.Registry.Actions, data.Actions...)
	//cfg.Registry.Workflows = append(cfg.Registry.Workflows, data.Workflows...)
	cfg.Registry.Executables = append(cfg.Registry.Executables, data.Executables...)

	// internal catalog
	internalCatalog := builtincatalog.InternalCatalog()
	cfg.Registry.Actions = append(cfg.Registry.Actions, internalCatalog.Actions...)
	cfg.Registry.Workflows = append(cfg.Registry.Workflows, internalCatalog.Workflows...)
	cfg.Registry.Executables = append(cfg.Registry.Executables, internalCatalog.Executables...)

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
