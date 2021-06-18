package api

import (
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
)

// ActionDetails holds details about the action
type ActionDetails struct {
	Stage            string
	Name             string
	Version          string
	UsedTools        []string
	ToolDependencies map[string]string
}

// ActionStep is the interface that needs to be implemented by all builtin actions
type ActionStep interface {
	// GetDetails retrieves information about the action
	GetDetails(ctx ActionExecutionContext) ActionDetails
	// Check evaluates if the action should be executed or not
	Check(ctx ActionExecutionContext) bool
	// Execute runs the action
	Execute(ctx ActionExecutionContext, state *ActionStateContext) error
}

// ActionExecutionContext holds runtime information for the actions
type ActionExecutionContext struct {
	// Paths holds the path configuration
	Paths config.PathConfig

	// ProjectDir holds the project directory
	ProjectDir string

	// WorkDir holds the current working directory
	WorkDir string

	// Config holds the yaml configuration passed to this action
	Config string

	// Args holds the arguments passed to the action
	Args []string

	// Env contains the normalized environment
	Env map[string]string

	// MachineEnv contains the full environment
	MachineEnv map[string]string

	// Parallelization defines how many tasks can be run in parallel inside of a action
	Parallelization int

	// Modules contains the project modules
	Modules []*analyzerapi.ProjectModule
}

// ActionStateContext holds state information about executed actions / results (ie. generated artifacts)
type ActionStateContext struct {
	// Version of the serialized action state
	Version int `json:"version"`
}

// CoverageReport contains a generic coverage report
type CoverageReport struct {
	Language string
	Percent float64
}

var BuiltinActions = make(map[string]ActionStep)

// RegisterBuiltinAction registers a builtin action
func RegisterBuiltinAction(action ActionStep) {
	ctx := ActionExecutionContext{}
	BuiltinActions[action.GetDetails(ctx).Name] = action
}
