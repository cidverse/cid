package golang

import (
	"github.com/cidverse/cid/pkg/common/config"
)

var Config = struct {
	Paths config.PathConfig
	GoLang struct {
		Platform []struct {
			Goos string `required:"true"`
			Goarch string `required:"true"`
		}
	}
}{}

func loadConfig(projectDirectory string) {
	_ = config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}