package node

import (
	"github.com/cidverse/cid/pkg/common/config"
)

var Config = struct {
	Paths config.PathConfig
}{}

func loadConfig(projectDirectory string) {
	_ = config.LoadConfigurationFile(&Config, projectDirectory+"/cid.yml")
}