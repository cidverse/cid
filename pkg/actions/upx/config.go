package upx

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/jinzhu/configor"
)

var Config = struct {
	Paths api.PathConfig
}{}

func loadConfig(projectDirectory string) {
	configor.Load(&Config, projectDirectory + "/mpi.yml")
}