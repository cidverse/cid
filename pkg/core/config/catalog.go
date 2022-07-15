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
)

type Action struct {
	Repository  string         `required:"true" yaml:"repository"`
	Name        string         `required:"true" yaml:"name"`
	Enabled     bool           `default:"true" yaml:"enabled,omitempty"`
	Type        ActionType     `required:"true" yaml:"type"`
	Description string         `yaml:"description"`
	Version     string         `yaml:"version"`
	Scope       ActionScope    `required:"true" yaml:"scope"`
	Rules       []WorkflowRule `yaml:"rules,omitempty"`
}
