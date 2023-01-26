package registry

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Source struct {
	URL string `yaml:"url"`
}

func getUserConfigDirectory() string {
	if runtime.GOOS == "windows" {
		cacheDir, _ := os.UserCacheDir()
		dir := filepath.Join(cacheDir, "cid")
		_ = os.MkdirAll(dir, os.ModePerm)

		return dir
	} else {
		homeDir, _ := os.UserHomeDir()
		dir := filepath.Join(homeDir, ".cache", "cid")
		_ = os.MkdirAll(dir, os.ModePerm)

		return dir
	}
}

func LoadSources() map[string]Source {
	sources := make(map[string]Source)
	file := filepath.Join(getUserConfigDirectory(), "repositories.yaml")

	// file doesn't exist yet, init with main repo
	if _, err := os.Stat(file); os.IsNotExist(err) {
		sources["central"] = Source{URL: "https://raw.githubusercontent.com/cidverse/registry/main/cid-index.yaml"}
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

func LoadRegistries() Config {
	var cfg Config

	sources := LoadSources()
	for name, _ := range sources {
		file := filepath.Join(getUserConfigDirectory(), "repo.d", name+".yaml")

		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Warn().Str("file", file).Msg("cache for registry is missing, please run `cid registry update`")
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

func saveSources(data map[string]Source) {
	file := filepath.Join(getUserConfigDirectory(), "repositories.yaml")

	out, err := yaml.Marshal(data)
	if err != nil {
		return
	}

	err = os.WriteFile(file, out, os.ModePerm)
	if err != nil {
		log.Fatal().Str("file", file).Msg("failed to update registries")
	}
}

func AddRegistry(name string, url string) {
	sources := LoadSources()
	sources[name] = Source{URL: url}
	saveSources(sources)
}

func RemoveRegistry(name string) {
	sources := LoadSources()
	delete(sources, name)
	saveSources(sources)
}

func UpdateAllRegistries() {
	sources := LoadSources()
	for name, source := range sources {
		UpdateRegistry(name, source)
	}
}
func UpdateRegistry(name string, source Source) {
	dir := filepath.Join(getUserConfigDirectory(), "repo.d")
	_ = os.MkdirAll(dir, os.ModePerm)

	// download
	file := filepath.Join(dir, name+".yaml")
	client := resty.New()
	resp, err := client.R().
		SetOutput(file).
		Get(source.URL)
	if err != nil {
		log.Fatal().Err(err).Str("url", source.URL).Msg("failed to fetch registry index")
	} else if resp.IsError() {
		log.Fatal().Str("url", source.URL).Str("response_status", resp.Status()).Msg("failed to fetch registry index")
	}
}
