package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Source struct {
	URI       string `yaml:"uri"`
	AddedAt   string `yaml:"added_at"`
	UpdatedAt string `yaml:"updated_at"`
	SHA256    string `yaml:"sha256"`
}

func LoadSources() map[string]*Source {
	sources := make(map[string]*Source)
	file := filepath.Join(util.CIDConfigDir(), "repositories.yaml")

	// file doesn't exist yet, init with main repo
	if _, err := os.Stat(file); os.IsNotExist(err) {
		sources["cid"] = &Source{URI: "https://raw.githubusercontent.com/cidverse/catalog/main/cid-index.yaml", AddedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339), SHA256: ""}
		return sources
	}

	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("failed to read registries")
	}

	err = yaml.Unmarshal(content, &sources)
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("failed to read registries")
	}

	return sources
}

func LoadCatalogs(sources map[string]*Source) Config {
	var cfg Config
	for name := range sources {
		file := filepath.Join(util.CIDConfigDir(), "repo.d", name+".yaml")

		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Warn().Str("file", file).Msg("cache for registry is missing, please run `cid catalog update`")
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			log.Error().Str("file", file).Msg("failed to read file")
			continue
		}

		var fileCfg Config
		err = yaml.Unmarshal(content, &fileCfg)
		if err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("failed to read registries")
		}

		// set repository
		for i := 0; i < len(fileCfg.ContainerImages); i++ {
			fileCfg.ContainerImages[i].Repository = name
		}
		for i := 0; i < len(fileCfg.Actions); i++ {
			fileCfg.Actions[i].Repository = name
		}
		for i := 0; i < len(fileCfg.Workflows); i++ {
			fileCfg.Workflows[i].Repository = name
		}

		// append
		cfg.ContainerImages = append(cfg.ContainerImages, fileCfg.ContainerImages...)
		cfg.Actions = append(cfg.Actions, fileCfg.Actions...)
		cfg.Workflows = append(cfg.Workflows, fileCfg.Workflows...)
	}

	return cfg
}

func saveSources(data map[string]*Source) {
	file := filepath.Join(util.CIDConfigDir(), "repositories.yaml")

	out, err := yaml.Marshal(data)
	if err != nil {
		return
	}

	err = os.WriteFile(file, out, os.ModePerm)
	if err != nil {
		log.Fatal().Str("file", file).Msg("failed to update registries")
	}
}

func AddCatalog(name string, url string) {
	sources := LoadSources()
	sources[name] = &Source{URI: url, AddedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}
	saveSources(sources)
}

func RemoveCatalog(name string) {
	sources := LoadSources()
	delete(sources, name)
	saveSources(sources)
}

func UpdateAllCatalogs() {
	sources := LoadSources()
	for name, source := range sources {
		UpdateCatalog(name, source)
		source.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	saveSources(sources)
}
func UpdateCatalog(name string, source *Source) {
	dir := filepath.Join(util.CIDConfigDir(), "repo.d")
	_ = os.MkdirAll(dir, os.ModePerm)

	// download
	file := filepath.Join(dir, name+".yaml")
	client := resty.New()
	resp, err := client.R().
		SetOutput(file).
		Get(source.URI)
	if err != nil {
		log.Fatal().Err(err).Str("uri", source.URI).Msg("failed to fetch registry index")
	} else if resp.IsError() {
		log.Fatal().Str("uri", source.URI).Str("response_status", resp.Status()).Msg("failed to fetch registry index")
	}

	// sha256 hash
	fileHash, hashErr := hash.SHA256Hash(bytes.NewReader(resp.Body()))
	if hashErr != nil {
		log.Fatal().Err(hashErr).Str("uri", source.URI).Msg("failed to calculate catalog hash")
	}
	source.SHA256 = fileHash
}
