package workflowrun

import (
	"encoding/json"
	"path/filepath"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

func getState(ctx *api.ActionExecutionContext) api.ActionStateContext {
	stateFile := filepath.Join(ctx.Paths.Artifact, "state.json")
	state := api.ActionStateContext{
		Version: 1,
		Modules: ctx.Modules,
	}
	if filesystem.FileExists(stateFile) {
		stateContent, stateContentErr := filesystem.GetFileContent(stateFile)
		if stateContentErr == nil {
			err := json.Unmarshal([]byte(stateContent), &state)
			if err != nil {
				log.Debug().Err(err).Str("file", stateFile).Msg("failed to restore state")
			}
		}
	}

	return state
}

func persistState(ctx *api.ActionExecutionContext, state api.ActionStateContext) {
	stateFile := filepath.Join(ctx.Paths.Artifact, "state.json")
	stateOut, err := json.Marshal(state)
	if err != nil {
		log.Warn().Err(err).Str("file", stateFile).Msg("failed to store state")
	} else {
		_ = filesystem.RemoveFile(stateFile)

		storeErr := filesystem.SaveFileContent(stateFile, string(stateOut))
		if storeErr != nil {
			log.Warn().Err(storeErr).Str("file", stateFile).Msg("failed to store state")
		}
	}
}
