package python

import (
	"github.com/cidverse/x/pkg/common/config"
)

var Config = struct {
	Paths config.PathConfig
}{}

func loadConfig(projectDirectory string) {
	config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}