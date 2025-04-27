package api

import (
	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/normalizeci/pkg/envstruct"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	"log/slog"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/redact"
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
	Paths           config.PathConfig            // Paths holds the path configuration
	ProjectDir      string                       // ProjectDir holds the project directory
	WorkDir         string                       // WorkDir holds the current working directory
	NCI             nci.Spec                     // NCI contains the NCI spec
	Config          interface{}                  // Config holds the json configuration passed to this action
	Args            []string                     // Args holds the arguments passed to the action
	Env             map[string]string            // Env contains the full environment
	ActionEnv       map[string]string            // ActionEnv contains the environment that is visible to the action
	Parallelization int                          // Parallelization defines how many tasks can be run in parallel inside an action
	CurrentUser     user.User                    // CurrentUser holds information about the user running this process
	Modules         []*analyzerapi.ProjectModule // Modules contains the project modules
	CurrentModule   *analyzerapi.ProjectModule   // CurrentModule contains the module that is currently being build
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
func GetActionContext(modules []*analyzerapi.ProjectModule, projectDir string, env map[string]string, access catalog.ActionAccess) ActionExecutionContext {
	actionEnv := make(map[string]string)

	// user
	currentUser, _ := user.Current()

	// only pass allowed env variables
	for k, v := range env {
		if strings.HasPrefix(k, "NCI_") {
			actionEnv[k] = v
			continue
		}

		if len(access.Environment) > 0 {
			for _, envAccess := range access.Environment {
				if envAccess.Pattern == true && regexp.MustCompile(envAccess.Name).MatchString(k) {
					actionEnv[k] = v

					if envAccess.Secret {
						redact.Redact(v)
					}
				} else if envAccess.Name == k {
					actionEnv[k] = v

					if envAccess.Secret {
						redact.Redact(v)
					}
				}
			}
		}
	}

	var nciSpec nci.Spec
	err := envstruct.EnvMapToStruct(&nciSpec, env)
	if err != nil {
		slog.With("err", err).Error("Failed to unmarshal nci spec in action context")
	}
	return ActionExecutionContext{
		Paths: config.PathConfig{
			Artifact: filepath.Join(projectDir, ".dist"),
			Temp:     filepath.Join(projectDir, ".tmp"),
			Cache:    "",
		},
		ProjectDir:      projectDir,
		WorkDir:         filesystem.WorkingDirOrPanic(),
		NCI:             nciSpec,
		Config:          "",
		Args:            nil,
		Env:             env,
		ActionEnv:       actionEnv,
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
