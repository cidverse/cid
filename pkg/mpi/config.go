package mpi

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/PhilippHeuer/cid/pkg/common/config"
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
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{}
}

func loadConfig(projectDirectory string) {
	config.LoadConfigurationFile(&Config, projectDirectory + "/cid.yml")
}
