package java

import (
	"github.com/qubid/x/pkg/common/config"
)

var Config = struct {
	Paths config.PathConfig
}{}

func loadConfig(projectDirectory string) {
	config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}