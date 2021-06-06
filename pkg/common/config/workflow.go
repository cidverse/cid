package config

type WorkflowStage struct {
	Name string
	Rules []WorkflowRule
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{} `yaml:"config,omitempty"`
}

type WorkflowRule struct {
	Expression string
}
