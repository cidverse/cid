package catalog

import (
	"strings"
)

type Action struct {
	Repository string          `yaml:"repository,omitempty" json:"repository,omitempty"`
	URI        string          `yaml:"uri" json:"uri"` // URI is a unique absolute identifier for the action
	Type       ActionType      `required:"true" yaml:"type" json:"type"`
	Container  ContainerAction `yaml:"container,omitempty" json:"container,omitempty"` // Container contains the configuration for containerized actions
	Version    string          `yaml:"version,omitempty" json:"version,omitempty"`
	Metadata   ActionMetadata  `yaml:"metadata" json:"metadata"`
}

type ActionMetadata struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Documentation string            `json:"documentation,omitempty"`
	Category      string            `json:"category"`
	Scope         ActionScope       `json:"scope"`
	Links         map[string]string `json:"links,omitempty"`
	Rules         []WorkflowRule    `json:"rules,omitempty"`  // Rules define conditions that must be met for the action to be executed
	Access        ActionAccess      `json:"access,omitempty"` // Access defines resources that the action may access
	Input         ActionInput       `json:"input,omitempty"`  // Input defines the inputs that the action may consume
	Output        ActionOutput      `json:"output,omitempty"` // Output defines the outputs that the action may produce
}

type ActionScope string

const (
	ActionScopeProject ActionScope = "project"
	ActionScopeModule  ActionScope = "module"
)

type ActionAccess struct {
	Environment []ActionAccessEnv        `json:"env,omitempty"`         // Environment variables that the action may access during execution
	Executables []ActionAccessExecutable `json:"executables,omitempty"` // Executables that the action may invoke during execution
	Network     []ActionAccessNetwork    `json:"network,omitempty"`     // Network access that the action may use during execution
}

type ActionAccessEnv struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Pattern     bool   `json:"pattern,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"` // Secret indicates that the environment variable holds a secret and should be redacted
}

type ActionAccessExecutable struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
}

type ActionAccessNetwork struct {
	Host string `json:"host"`
}

type ActionInput struct {
	Artifacts []ActionArtifactType `json:"artifacts,omitempty"`
}

type ActionOutput struct {
	Artifacts []ActionArtifactType `json:"artifacts,omitempty"`
}

type ActionArtifactType struct {
	Type          string `json:"type"`             // Type, e.g. "report", "binary"
	Format        string `json:"format,omitempty"` // Format, e.g. "sarif"
	FormatVersion string `json:"format_version,omitempty"`
}

func (a ActionArtifactType) Key() string {
	var parts []string
	if a.Type != "" {
		parts = append(parts, a.Type)
	}
	if a.Format != "" {
		parts = append(parts, a.Format)
	}
	if a.FormatVersion != "" {
		parts = append(parts, a.FormatVersion)
	}

	return strings.Join(parts, ":")
}

type ActionType string

const (
	ActionTypeContainer    ActionType = "container"
	ActionTypeGitHubAction ActionType = "githubaction"
)

type ContainerAction struct {
	Image   string       `json:"image"`   // Image is the full image reference including the registry
	Command string       `json:"command"` // Command is the command that should be executed in the container image to start the action.
	Certs   []ImageCerts `json:"certs,omitempty"`
}
