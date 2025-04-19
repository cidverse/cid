package state

import (
	"time"

	"github.com/cidverse/repoanalyzer/analyzerapi"
)

// ActionArtifact contains information about generated artifacts
type ActionArtifact struct {
	BuildID       string `json:"build_id"`
	JobID         string `json:"job_id"`
	StepSlug      string `json:"step_slug"`
	ArtifactID    string `json:"id"`
	Module        string `json:"module"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Format        string `json:"format"`
	FormatVersion string `json:"format_version"`
	SHA256        string `json:"sha256"`
}

// AuditEvents contains information about all steps that were part of the build and deployment process
type AuditEvents struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Payload   map[string]string `json:"payload"`
}

// ActionStateContext holds state information about executed actions / results (ie. generated artifacts)
type ActionStateContext struct {
	// Version of the serialized action state
	Version int `json:"version"`

	// Modules contains the project modules
	Modules []*analyzerapi.ProjectModule

	// Artifacts
	Artifacts map[string]ActionArtifact `json:"artifacts"`

	// Steps holds a list of all steps that were part of the pipeline
	AuditLog []AuditEvents `json:"audit_events"`
}
