package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/hash"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var repositoryConfigFile = filepath.Join(util.CIDConfigDir(), "repositories.json")

type Source struct {
	URI       string   `json:"uri"`
	AddedAt   string   `json:"added_at"`
	UpdatedAt string   `json:"updated_at"`
	SHA256    string   `json:"sha256"`
	Filter    []string `json:"filter"`
}

func LoadSources() map[string]*Source {
	sources := make(map[string]*Source)

	// file doesn't exist yet, init with main repo
	if _, err := os.Stat(repositoryConfigFile); os.IsNotExist(err) {
		sources["cid"] = &Source{URI: "https://raw.githubusercontent.com/cidverse/catalog/main/cid-index.yaml", AddedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339), SHA256: ""}
		return sources
	}

	content, err := os.ReadFile(repositoryConfigFile)
	if err != nil {
		log.Fatal().Err(err).Str("file", repositoryConfigFile).Msg("failed to read registries")
	}

	err = json.Unmarshal(content, &sources)
	if err != nil {
		log.Fatal().Err(err).Str("file", repositoryConfigFile).Msg("failed to read registries")
	}

	return sources
}

func LoadCatalogs(sources map[string]*Source) Config {
	var cfg Config
	for name := range sources {
		file := filepath.Join(util.CIDConfigDir(), "repo.d", name+".json")

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
		err = json.Unmarshal(content, &fileCfg)
		if err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("failed to read registries")
		}

		// set repository
		for i := 0; i < len(fileCfg.Actions); i++ {
			fileCfg.Actions[i].Repository = name
		}
		for i := 0; i < len(fileCfg.Workflows); i++ {
			fileCfg.Workflows[i].Repository = name
		}

		// append
		cfg.Actions = append(cfg.Actions, fileCfg.Actions...)
		cfg.Workflows = append(cfg.Workflows, fileCfg.Workflows...)
	}

	return cfg
}

func saveSources(data map[string]*Source) {
	out, err := json.Marshal(data)
	if err != nil {
		return
	}

	err = os.WriteFile(repositoryConfigFile, out, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Str("file", repositoryConfigFile).Msg("failed to update registries")
	}
}

func AddCatalog(name string, url string, types []string) {
	sources := LoadSources()
	sources[name] = &Source{URI: url, Filter: types, AddedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}
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

func UpdateCatalog(name string, source *Source) error {
	dir := filepath.Join(util.CIDConfigDir(), "repo.d")
	file := filepath.Join(dir, name+".json")

	if strings.HasPrefix(source.URI, "oci://") {
		return updateCatalogOCI(file, source)
	} else {
		return updateCatalogFile(file, source)
	}
}

func updateCatalogOCI(file string, source *Source) error {
	// get metadata from oci image
	ociImage := strings.TrimPrefix(source.URI, "oci://")

	// configure container
	containerExec := containerruntime.Container{
		Image:   ociImage,
		Command: "central cid-metadata",
		User:    util.GetContainerUser(),
	}
	containerCmd, err := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if err != nil {
		return err
	}
	var outputBuffer bytes.Buffer
	cmd, err := shellcommand.PrepareCommand(containerCmd, runtime.GOOS, "", true, nil, "", nil, &outputBuffer, os.Stderr)
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	// parse json
	var actionMetadata []ActionMetadata
	err = json.Unmarshal(outputBuffer.Bytes(), &actionMetadata)

	// preprocess data
	data := Config{
		Actions: make([]Action, 0),
	}
	for _, am := range actionMetadata {
		data.Actions = append(data.Actions, Action{
			Repository: source.URI,
			Type:       ActionTypeContainer,
			Container: ContainerAction{
				Image:   ociImage,
				Command: "central run " + am.Name,
				Certs:   nil,
			},
			Version:  ociImage,
			Metadata: am,
		})
	}
	output, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal catalog data for %s: %w", source.URI, err)
	}

	// persist
	err = os.WriteFile(file, output, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write catalog file for %s: %w", source.URI, err)
	}

	// sha256 hash
	fileHash, hashErr := hash.SHA256Hash(bytes.NewReader(output))
	if hashErr != nil {
		return fmt.Errorf("failed to calculate catalog hash for %s: %w", source.URI, hashErr)
	}
	source.SHA256 = fileHash

	return nil
}

func updateCatalogFile(file string, source *Source) error {
	// download
	client := resty.New()
	resp, err := client.R().
		Get(source.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch registry index for %s: %w", source.URI, err)
	} else if resp.IsError() {
		return fmt.Errorf("failed to fetch registry index for %s: %s", source.URI, resp.Status())
	}

	// get content
	content := resp.Body()

	// transform yaml to json if yaml
	config := Config{}
	if strings.HasSuffix(source.URI, ".yaml") || strings.HasSuffix(source.URI, ".yml") {
		err = yaml.Unmarshal(content, &config)
		if err != nil {
			return fmt.Errorf("failed to parse yaml for %s: %w", source.URI, err)
		}
	}

	// filter
	if len(source.Filter) > 0 && !slices.Contains(source.Filter, "actions") {
		config.Actions = nil
	}
	if len(source.Filter) > 0 && !slices.Contains(source.Filter, "workflows") {
		config.Workflows = nil
	}

	// persist
	content, err = json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal catalog data for %s: %w", source.URI, err)
	}
	err = os.WriteFile(file, content, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write catalog file for %s: %w", source.URI, err)
	}

	// sha256 hash
	fileHash, hashErr := hash.SHA256Hash(bytes.NewReader(content))
	if hashErr != nil {
		return fmt.Errorf("failed to calculate catalog hash for %s: %w", source.URI, hashErr)
	}
	source.SHA256 = fileHash

	return nil
}
