package config

type Catalog struct {
	Actions []Action `json:"actions"`
}

type ActionScope string

const (
	ActionScopeProject ActionScope = "project"
	ActionScopeModule  ActionScope = "module"
)

type ActionType string

const (
	ActionTypeBuiltinGolang ActionType = "builtin-golang"
	ActionTypeContainer     ActionType = "container"
)

type ActionAccess struct {
	Env []string `yaml:"env"`
}

type ContainerAction struct {
	Image   string `yaml:"image"`   // Image is the full image reference including the registry
	Command string `yaml:"command"` // Command is the command that should be executed in the container image to start the action.
}

type Action struct {
	Repository  string          `required:"true" yaml:"repository"`
	Name        string          `required:"true" yaml:"name"`
	Enabled     bool            `default:"true" yaml:"enabled,omitempty"`
	Type        ActionType      `required:"true" yaml:"type"`
	Container   ContainerAction `yaml:"container,omitempty"` // Container contains the configuration for containerized actions
	Description string          `yaml:"description"`
	Version     string          `yaml:"version"`
	Scope       ActionScope     `required:"true" yaml:"scope"`
	Rules       []WorkflowRule  `yaml:"rules,omitempty"`
	Access      ActionAccess    `yaml:"access"`
}
