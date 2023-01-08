package app

import (
	"github.com/cidverse/cid/pkg/core/config"
)

func Load(projectDirectory string) *config.CIDConfig {
	// load configuration for the current project
	return config.LoadConfig(projectDirectory)
}
