package config

type WorkflowStage struct {
	Name string
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{} `yaml:"config,omitempty"`
}

// FindWorkflowStages finds all relevant stages for the current context (branch, tag, ...)
func FindWorkflowStages(projectDir string, env map[string]string) []WorkflowStage {
	return Config.Stages
}
