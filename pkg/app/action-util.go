package app

import (
	"github.com/cidverse/cid/pkg/actions/golang"
	"github.com/cidverse/cid/pkg/common/config"
)

func Load(projectDirectory string) {
	// load configuration for the current project
	config.LoadConfig(projectDirectory)

	// dependency detection
	// this will try to discover version constraints from the projects automatically
	dependencyDetectors := [...]map[string]string{
		golang.GetDependencies(projectDirectory),
	}

	for _, dep := range dependencyDetectors {
		for key, version := range dep {
			_, present := config.Config.Dependencies[key]
			if !present {
				config.Config.Dependencies[key] = version
			}
		}
	}
}
