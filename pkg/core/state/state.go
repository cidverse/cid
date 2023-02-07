package state

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type ActionArtifactType string

const (
	ActionArtifactTypeBinary ActionArtifactType = "binary"
	ActionArtifactTypeReport ActionArtifactType = "report"
)

// ActionArtifact contains information about generated artifacts
type ActionArtifact struct {
	BuildID       string             `json:"build_id"`
	JobID         string             `json:"job_id"`
	ArtifactID    string             `json:"artifact_id"`
	Module        string             `json:"module"`
	Type          ActionArtifactType `json:"type"`
	Name          string             `json:"name"`
	Format        string             `json:"format"`
	FormatVersion string             `json:"format_version"`
}

// ActionStateContext holds state information about executed actions / results (ie. generated artifacts)
type ActionStateContext struct {
	// Version of the serialized action state
	Version int `json:"version"`

	// Modules contains the project modules
	Modules []*analyzerapi.ProjectModule

	// Artifacts
	Artifacts map[string]ActionArtifact `json:"artifacts"`
}
