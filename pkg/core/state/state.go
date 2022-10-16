package state

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

// ActionStateContext holds state information about executed actions / results (ie. generated artifacts)
type ActionStateContext struct {
	// Version of the serialized action state
	Version int `json:"version"`

	// Modules contains the project modules
	Modules []*analyzerapi.ProjectModule
}
