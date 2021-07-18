package app

import (
	_ "github.com/cidverse/cid/pkg/actions/changelog"
	_ "github.com/cidverse/cid/pkg/actions/container"
	_ "github.com/cidverse/cid/pkg/actions/fossa"
	_ "github.com/cidverse/cid/pkg/actions/gitguardian"
	_ "github.com/cidverse/cid/pkg/actions/gitleaks"
	_ "github.com/cidverse/cid/pkg/actions/golang"
	_ "github.com/cidverse/cid/pkg/actions/helm"
	_ "github.com/cidverse/cid/pkg/actions/hugo"
	_ "github.com/cidverse/cid/pkg/actions/java"
	_ "github.com/cidverse/cid/pkg/actions/node"
	_ "github.com/cidverse/cid/pkg/actions/owaspdepcheck"
	_ "github.com/cidverse/cid/pkg/actions/python"
	_ "github.com/cidverse/cid/pkg/actions/repo"
	_ "github.com/cidverse/cid/pkg/actions/sonarqube"
	_ "github.com/cidverse/cid/pkg/actions/trivy"
	_ "github.com/cidverse/cid/pkg/actions/upx"
	"github.com/cidverse/cid/pkg/common/config"
)

func Load(projectDirectory string) {
	// load configuration for the current project
	config.LoadConfig(projectDirectory)
}
