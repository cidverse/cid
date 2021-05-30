package hugo

import (
	"github.com/cidverse/x/pkg/common/config"
)

var Config = struct {
	Paths config.PathConfig
	Hugo struct {
		Platform []struct {
			Goos string `required:"true"`
			Goarch string `required:"true"`
		}
	}
}{}

func loadConfig(projectDirectory string) {
	_ = config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}