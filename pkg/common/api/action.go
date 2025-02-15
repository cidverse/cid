package api

import (
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/normalizeci/pkg/normalizer/api"
	"github.com/samber/lo"

	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

const DefaultParallelization = 10

// ActionDetails holds details about the action
type ActionDetails struct {
	Name             string
	Version          string
	UsedTools        []string
	ToolDependencies map[string]string
}

// ActionStep is the interface that needs to be implemented by all builtin actions
type ActionStep interface {
	// GetDetails retrieves information about the action
	GetDetails(ctx *ActionExecutionContext) ActionDetails

	// Execute runs the action
	Execute(ctx *ActionExecutionContext, localState *state.ActionStateContext) error
}

// ActionExecutionContext holds runtime information for the actions
type ActionExecutionContext struct {
	// Paths holds the path configuration
	Paths config.PathConfig

	// ProjectDir holds the project directory
	ProjectDir string

	// WorkDir holds the current working directory
	WorkDir string

	// Config holds the json configuration passed to this action
	Config string

	// Args holds the arguments passed to the action
	Args []string

	// Env contains the environment that is visible to the action
	Env map[string]string

	// Parallelization defines how many tasks can be run in parallel inside an action
	Parallelization int

	// CurrentUser holds information about the user running this process
	CurrentUser user.User

	// Modules contains the project modules
	Modules []*analyzerapi.ProjectModule

	// CurrentModule contains the module that is currently being build
	CurrentModule *analyzerapi.ProjectModule
}

// CoverageReport contains a generic coverage report
type CoverageReport struct {
	Language string
	Percent  float64
}

var BuiltinActions = make(map[string]ActionStep)

// RegisterBuiltinAction registers a builtin action
func RegisterBuiltinAction(action ActionStep) {
	ctx := ActionExecutionContext{}
	BuiltinActions[action.GetDetails(&ctx).Name] = action
}

// GetActionContext gets the action context, this operation is expensive and should only be called once per execution
func GetActionContext(modules []*analyzerapi.ProjectModule, projectDir string, env map[string]string, access *catalog.ActionAccess) ActionExecutionContext {
	actionEnv := make(map[string]string)

	// user
	currentUser, _ := user.Current()

	// only pass allowed env variables
	fullEnv := lo.Assign(env, api.GetMachineEnvironment())
	for k, v := range fullEnv {
		if strings.HasPrefix(k, "NCI_") {
			actionEnv[k] = v
			continue
		}

		if access != nil && len(access.Environment) > 0 {
			for _, envAccess := range access.Environment {
				if envAccess.Pattern == true && regexp.MustCompile(envAccess.Name).MatchString(k) {
					actionEnv[k] = v
				} else if envAccess.Name == k {
					actionEnv[k] = v
				}
			}
		}
	}

	return ActionExecutionContext{
		Paths: config.PathConfig{
			Artifact: filepath.Join(projectDir, ".dist"),
			Temp:     filepath.Join(projectDir, ".tmp"),
			Cache:    "",
		},
		ProjectDir:      projectDir,
		WorkDir:         filesystem.WorkingDirOrPanic(),
		Config:          "",
		Args:            nil,
		Env:             actionEnv,
		Parallelization: DefaultParallelization,
		CurrentUser:     *currentUser,
		Modules:         modules,
		CurrentModule:   nil,
	}
}

// MissingRequirement contains a record about a missing requirement for a action
type MissingRequirement struct {
	Message string
}
