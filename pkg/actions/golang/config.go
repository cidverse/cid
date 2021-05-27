package golang

import (
	"github.com/cidverse/x/pkg/common/config"
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
	config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}