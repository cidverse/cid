package catalog

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadFromDirectory(dir string) (*Config, error) {
	var data Config

	// list all yaml files in directory
	loadErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), "cid-index.yaml") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Parse yaml file
			var fileData Config
			err = yaml.Unmarshal(content, &fileData)
			if err != nil {
				return err
			}

			// merge actions
			if len(fileData.Actions) > 0 {
				for _, action := range fileData.Actions {
					data.Actions = append(data.Actions, action)
				}
			}

			// merge images
			if len(fileData.ContainerImages) > 0 {
				for _, image := range fileData.ContainerImages {
					data.ContainerImages = append(data.ContainerImages, image)
				}
			}

			// merge workflows
			if len(fileData.Workflows) > 0 {
				for _, workflow := range fileData.Workflows {
					data.Workflows = append(data.Workflows, workflow)
				}
			}
		}
		return nil
	})
	if loadErr != nil {
		return nil, loadErr
	}

	return &data, nil
}

func LoadFromFile(file string) (*Config, error) {
	// load file
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// parse
	var data Config
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	// test
	return &data, nil
}

func SaveToFile(registry *Config, file string) error {
	// marshal
	data, err := yaml.Marshal(&registry)
	if err != nil {
		return err
	}

	// write to filesystem
	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
