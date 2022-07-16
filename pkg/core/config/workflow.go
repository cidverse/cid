package config

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
)

type WorkflowExpressionType string

const (
	WorkflowExpressionCEL WorkflowExpressionType = "cel"
)

type WorkflowRule struct {
	Type       WorkflowExpressionType `default:"cel" yaml:"type,omitempty"`
	Expression string                 `yaml:"expression,omitempty"`
}

type WorkflowAction struct {
	Id     string                     `required:"true" yaml:"id"`
	Rules  []WorkflowRule             `yaml:"rules,omitempty"`
	Config interface{}                `yaml:"config,omitempty"`
	Module *analyzerapi.ProjectModule `yaml:"-"`
}

type WorkflowStage struct {
	Name    string           `required:"true" yaml:"name,omitempty"`
	Rules   []WorkflowRule   `yaml:"rules,omitempty"`
	Actions []WorkflowAction `yaml:"actions,omitempty"`
}

type Workflow struct {
	Name   string          `required:"true" yaml:"name,omitempty"`
	Rules  []WorkflowRule  `yaml:"rules,omitempty"`
	Stages []WorkflowStage `yaml:"stages,omitempty"`
}

// FindWorkflow finds a workflow by name
func (c CIDConfig) FindWorkflow(name string) *Workflow {
	for _, w := range c.Workflows {
		if w.Name == name {
			return &w
		}
	}

	return nil
}

// FindAction finds a action by id
func (c CIDConfig) FindAction(name string) *Action {
	// exact match
	for _, a := range c.Catalog.Actions {
		if a.Repository+"/"+a.Name == name {
			return &a
		}
	}

	return nil
}

// ActionCount returns the total count of actions across all stages
func (w Workflow) ActionCount() int {
	actionCount := 0
	for _, s := range w.Stages {
		actionCount = actionCount + len(s.Actions)
	}

	return actionCount
}
