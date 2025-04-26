package builtinaction

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrybuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrytest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvtest"
)

// GetActionsMetadata returns a map of all actions with their metadata
func GetActionsMetadata() map[string]cidsdk.Action {
	sdk, _ := cidsdk.NewSDK()
	return GetActions(sdk)
}

// GetActions returns a map of all actions initialized with the given SDK
func GetActions(sdk *cidsdk.SDK) map[string]cidsdk.Action {
	// actions
	actions := []cidsdk.Action{
		// python-poetry
		poetrybuild.Action{Sdk: *sdk},
		poetrytest.Action{Sdk: *sdk},
		// python-uv
		uvbuild.Action{Sdk: *sdk},
		uvtest.Action{Sdk: *sdk},
	}

	// as map
	actionMap := make(map[string]cidsdk.Action, len(actions))
	for _, action := range actions {
		actionMap[action.Metadata().Name] = action
	}

	return actionMap
}
