package mpi

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/jinzhu/configor"
)

var Config = struct {
	Paths api.PathConfig
	Workflow []WorkflowStage
}{}

type WorkflowStage struct {
	Stage string
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string
}

func loadConfig(projectDirectory string) {
	configor.New(&configor.Config{ENVPrefix: "MPI"}).Load(&Config, projectDirectory + "/mpi.yml")
}
