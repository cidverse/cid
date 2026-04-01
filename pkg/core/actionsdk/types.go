package actionsdk

import "strings"

// Action is the common interface for all actions
type Action interface {
	Metadata() ActionMetadata
	Execute() error
}

type ActionMetadata struct {
	Name          string            `json:"name"`                   // Name is the name of the action
	Description   string            `json:"description"`            // Description is a short one-line description of the action
	Documentation string            `json:"documentation"`          // Documentation is a longer multi-line description of the action
	Category      string            `json:"category"`               // Category is the category of the action, e.g. "build", "test", "deploy"
	Scope         ActionScope       `json:"scope"`                  // Scope of the action, either "project" or "module"
	Links         map[string]string `json:"links,omitempty"`        // Links to additional documentation
	Rules         []ActionRule      `json:"rules,omitempty"`        // Rules define conditions that must be met for the action to be executed
	RunIfChanged  []string          `json:"runIfChanged,omitempty"` // RunIfChanged defines files that must be changed for the action to be executed
	Access        ActionAccess      `json:"access,omitempty"`       // Access defines resources that the action may access
	Input         ActionInput       `json:"input,omitempty"`        // Input defines the inputs that the action may consume
	Output        ActionOutput      `json:"output,omitempty"`       // Output defines the outputs that the action may produce
}

func (am *ActionMetadata) HasOutputWithTypeAndFormat(artifactType string, artifactFormat string) bool {
	for _, artifact := range am.Output.Artifacts {
		if artifact.Type == artifactType && artifact.Format == artifactFormat {
			return true
		}
	}

	return false
}

type ActionScope string

const (
	ActionScopeProject ActionScope = "project"
	ActionScopeModule  ActionScope = "module"
)

type ActionRule struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type ActionAccess struct {
	Environment []ActionAccessEnv        `json:"env,omitempty"`         // Environment variables that the action may access during execution
	Executables []ActionAccessExecutable `json:"executables,omitempty"` // Executables that the action may invoke during execution
	Network     []ActionAccessNetwork    `json:"network,omitempty"`     // Network access that the action may use during execution
	Resources   []ActionAccessResource   `json:"resources,omitempty"`   // Resources the action may access (e.g. releases, packages, ...)
}

type ActionAccessEnv struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Pattern     bool   `json:"pattern,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type ActionAccessExecutable struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
}

type ActionAccessNetwork struct {
	Host string `json:"host"`
}

type ActionAccessResource string

const (
	ResourceReleases ActionAccessResource = "releases"
	ResourcePackages ActionAccessResource = "packages"
)

type ActionInput struct {
	Artifacts []ActionArtifactType `json:"artifacts,omitempty"`
}

type ActionOutput struct {
	Artifacts []ActionArtifactType `json:"artifacts,omitempty"`
}

func (a ActionOutput) ContainsArtifactWithTypeAndFormat(artifactType string, artifactFormat string) bool {
	for _, artifact := range a.Artifacts {
		if artifact.Type == artifactType && artifact.Format == artifactFormat {
			return true
		}
	}
	return false
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
