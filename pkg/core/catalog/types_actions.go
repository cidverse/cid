package catalog

import (
	"fmt"

	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

type Action struct {
	Repository string          `yaml:"repository,omitempty" json:"repository,omitempty"`
	URI        string          `yaml:"uri" json:"uri"` // URI is a unique absolute identifier for the action
	Type       ActionType      `required:"true" yaml:"type" json:"type"`
	Container  ContainerAction `yaml:"container,omitempty" json:"container,omitempty"` // Container contains the configuration for containerized actions
	Version    string          `yaml:"version,omitempty" json:"version,omitempty"`
	Metadata   ActionMetadata  `yaml:"metadata" json:"metadata"`
}

// AuditPayload returns the audit log entry for executing this action. The uri is the digest-pinned
// identifier, reference carries the human readable tag for information only.
func (a Action) AuditPayload() map[string]string {
	return map[string]string{
		"action":    a.Repository + "/" + a.Metadata.Name + "@" + a.Version,
		"uri":       fmt.Sprintf("oci://%s", a.Container.Reference()),
		"reference": a.Container.NamedReference(),
	}
}

type ActionMetadata struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Documentation string                 `json:"documentation,omitempty"`
	Category      string                 `json:"category"`
	Scope         actionsdk.ActionScope  `json:"scope"`
	Links         map[string]string      `json:"links,omitempty"`
	Rules         []WorkflowRule         `json:"rules,omitempty"`        // Rules define conditions that must be met for the action to be executed
	RunIfChanged  []string               `json:"runIfChanged,omitempty"` // RunIfChanged defines files that must be changed for the action to be executed
	Access        actionsdk.ActionAccess `json:"access,omitempty"`       // Access defines resources that the action may access
	Input         actionsdk.ActionInput  `json:"input,omitempty"`        // Input defines the inputs that the action may consume
	Output        actionsdk.ActionOutput `json:"output,omitempty"`       // Output defines the outputs that the action may produce
}

type ActionType string

const (
	ActionTypeBuiltIn      ActionType = "builtin"
	ActionTypeContainer    ActionType = "container"
	ActionTypeGitHubAction ActionType = "githubaction"
)

type ContainerAction struct {
	Image   string       `json:"image"`            // Image is the full image reference including the registry
	Digest  string       `json:"digest,omitempty"` // Digest is the manifest digest the image tag was resolved to, used to pin execution
	Command string       `json:"command"`          // Command is the command that should be executed in the container image to start the action.
	Certs   []ImageCerts `json:"certs,omitempty"`
}

// Reference returns the digest-pinned image reference, falling back to the tag if no digest is known.
func (c ContainerAction) Reference() string {
	return executable.PinnedReference(c.Image, c.Digest)
}

// NamedReference returns the human readable tag reference of the image, empty for digest-only images.
func (c ContainerAction) NamedReference() string {
	return executable.NamedReference(c.Image)
}
