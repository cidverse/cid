package catalog

type Action struct {
	Repository string          `yaml:"repository,omitempty"`
	Type       ActionType      `required:"true" yaml:"type"`
	Container  ContainerAction `yaml:"container,omitempty"` // Container contains the configuration for containerized actions
	Version    string          `yaml:"version,omitempty"`
	Metadata   ActionMetadata  `yaml:"metadata"`
}

type ActionMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Scope       ActionScope       `json:"scope"`
	Links       map[string]string `json:"links,omitempty"`
	Rules       []WorkflowRule    `yaml:"rules,omitempty"`  // Rules define conditions that must be met for the action to be executed
	Access      ActionAccess      `json:"access,omitempty"` // Access defines resources that the action may access
}

type ActionScope string

const (
	ActionScopeProject ActionScope = "project"
	ActionScopeModule  ActionScope = "module"
)

type ActionAccess struct {
	Environment []ActionAccessEnv        `json:"env,omitempty"`         // Environment variables that the action may access during execution
	Executables []ActionAccessExecutable `json:"executables,omitempty"` // Executables that the action may invoke during execution
}

type ActionAccessEnv struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Pattern     bool   `json:"pattern,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type ActionAccessExecutable struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
}

type ActionType string

const (
	ActionTypeContainer    ActionType = "container"
	ActionTypeGitHubAction ActionType = "githubaction"
)

type ContainerAction struct {
	Image   string       `yaml:"image"`   // Image is the full image reference including the registry
	Command string       `yaml:"command"` // Command is the command that should be executed in the container image to start the action.
	Certs   []ImageCerts `yaml:"certs,omitempty"`
}
