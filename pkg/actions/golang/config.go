package golang

import "github.com/jinzhu/configor"

var GoConfig = struct {
	GoLang struct {
		Platform []struct {
			Goos string `required:"true"`
			Goarch string `required:"true"`
		}
	}
}{}

func loadConfig(projectDirectory string) {
	configor.Load(&GoConfig, projectDirectory + "/config.yml")
}