package app

import (
	_ "github.com/cidverse/cid/pkg/actions/container"
	_ "github.com/cidverse/cid/pkg/actions/golang"
	_ "github.com/cidverse/cid/pkg/actions/helm"
	_ "github.com/cidverse/cid/pkg/actions/java"
	_ "github.com/cidverse/cid/pkg/actions/repo"
	_ "github.com/cidverse/cid/pkg/actions/sonarqube"
	_ "github.com/cidverse/cid/pkg/actions/trivy"
	"github.com/cidverse/cid/pkg/core/config"
)

func Load(projectDirectory string) *config.CIDConfig {
	// load configuration for the current project
	return config.LoadConfig(projectDirectory)
}
