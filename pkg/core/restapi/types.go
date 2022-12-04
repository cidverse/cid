package restapi

import (
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type handlerConfig struct {
	projectDir          string
	containerProjectDir string
	modules             []*analyzerapi.ProjectModule
	currentModule       *analyzerapi.ProjectModule
	currentAction       *config.Action
	env                 map[string]string
	actionConfig        string
}

// apiError, see https://www.rfc-editor.org/rfc/rfc7807
type apiError struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}