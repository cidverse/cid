package config

import "github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"

type WorkflowStage struct {
	Name    string
	Rules   []WorkflowRule
	Actions []WorkflowAction
}

type WorkflowAction struct {
	// Name of the action
	Name string `required:"true"`

	// Type of the action, does determinate how a action is executed
	Type string `default:"builtin"`

	// Config holds custom configuration options for this action
	Config interface{} `yaml:"config,omitempty"`

	// Module is the module being build, this is set automatically
	Module *analyzerapi.ProjectModule
}

type WorkflowRule struct {
	Expression string
}
