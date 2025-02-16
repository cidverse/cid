package catalog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/executable"
	"gopkg.in/yaml.v3"
)

func LoadFromDirectory(dir string) (*Config, error) {
	data := Config{
		Actions:   nil,
		Workflows: nil,
		ExecutableDiscovery: &ExecutableDiscovery{
			ContainerDiscovery: executable.DiscoverContainerOptions{
				Packages: nil,
			},
		},
		Executables: nil,
	}

	// list all yaml files in directory
	loadErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
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

			// merge workflows
			if len(fileData.Workflows) > 0 {
				for _, workflow := range fileData.Workflows {
					data.Workflows = append(data.Workflows, workflow)
				}
			}

			// merge executable discovery
			if fileData.ExecutableDiscovery != nil && fileData.ExecutableDiscovery.ContainerDiscovery.Packages != nil {
				data.ExecutableDiscovery.ContainerDiscovery.Packages = append(data.ExecutableDiscovery.ContainerDiscovery.Packages, fileData.ExecutableDiscovery.ContainerDiscovery.Packages...)
			}
		}
		return nil
	})
	if loadErr != nil {
		return nil, loadErr
	}

	return &data, nil
}
