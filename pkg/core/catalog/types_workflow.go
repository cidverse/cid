package catalog

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
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
	ID     string                     `required:"true" yaml:"id"`
	Rules  []WorkflowRule             `yaml:"rules,omitempty"`
	Config interface{}                `yaml:"config,omitempty"`
	Module *analyzerapi.ProjectModule `yaml:"-"`
	Stage  string                     `yaml:"-"`
}

type WorkflowStage struct {
	Name    string           `required:"true" yaml:"name,omitempty"`
	Rules   []WorkflowRule   `yaml:"rules,omitempty"`
	Actions []WorkflowAction `yaml:"actions,omitempty"`
}

type Workflow struct {
	Repository  string          `yaml:"repository,omitempty"`
	Name        string          `required:"true" yaml:"name,omitempty"`
	Description string          `yaml:"description,omitempty"`
	Version     string          `yaml:"version,omitempty"`
	Rules       []WorkflowRule  `yaml:"rules,omitempty"`
	Stages      []WorkflowStage `yaml:"stages,omitempty"`
}

// ActionCount returns the total count of actions across all stages
func (w *Workflow) ActionCount() int {
	actionCount := 0
	for _, s := range w.Stages {
		actionCount += len(s.Actions)
	}

	return actionCount
}
