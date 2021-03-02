package hugo

import "github.com/jinzhu/configor"

var HugoConfig = struct {
	Hugo struct {
		Platform []struct {
			Goos string `required:"true"`
			Goarch string `required:"true"`
		}
	}
}{}

func loadConfig(projectDirectory string) {
	configor.Load(&HugoConfig, projectDirectory + "/mpi.yml")
}