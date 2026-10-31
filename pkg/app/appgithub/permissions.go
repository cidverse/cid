package appgithub

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/plangenerate"
)

// WorkflowPermission is a single GITHUB_TOKEN scope rendered for a job.
type WorkflowPermission struct {
	Name   string // Scope name, e.g. "contents"
	Level  string // Access level, e.g. "read" or "write"
	Reason string // Comment explaining why the scope is required
}

// jobPermissionOrder defines the stable order in which scopes are rendered.
var jobPermissionOrder = []string{
	"actions",
	"contents",
	"id-token",
	"security-events",
	"packages",
	"pull-requests",
}

// jobPermissions returns the GITHUB_TOKEN scopes for a job.
//
// Every job starts with the baseline scopes required to check out the repository and query workflow metadata.
func jobPermissions(step plangenerate.Step, workflowType string) []WorkflowPermission {
	permissions := map[string]WorkflowPermission{
		"actions":  {Name: "actions", Level: "read", Reason: "required: query workflow run metadata"},
		"contents": {Name: "contents", Level: "read", Reason: "required: checkout / repository metadata"},
	}

	// write scopes are not granted on pull requests, where secrets and releases are unavailable.
	if workflowType != "pull-request" {
		permissions["id-token"] = WorkflowPermission{Name: "id-token", Level: "write", Reason: "required: provenance signing"}

		for _, resource := range step.Access.Resources {
			switch resource {
			case actionsdk.ResourceSecurityEvents:
				permissions["security-events"] = WorkflowPermission{Name: "security-events", Level: "write", Reason: "required: upload security events"}
			case actionsdk.ResourceReleases:
				permissions["contents"] = WorkflowPermission{Name: "contents", Level: "write", Reason: "required: create release"}
			case actionsdk.ResourcePackages:
				permissions["packages"] = WorkflowPermission{Name: "packages", Level: "write", Reason: "required: publication"}
			}
		}
	}

	if step.Name == "github-pr-comment" {
		permissions["pull-requests"] = WorkflowPermission{Name: "pull-requests", Level: "write", Reason: "post pr comment"}
	}

	ordered := make([]WorkflowPermission, 0, len(permissions))
	for _, name := range jobPermissionOrder {
		if permission, ok := permissions[name]; ok {
			ordered = append(ordered, permission)
		}
	}

	return ordered
}
