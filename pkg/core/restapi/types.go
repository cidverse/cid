package restapi

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type handlerConfig struct {
	projectDir    string
	modules       []*analyzerapi.ProjectModule
	currentModule *analyzerapi.ProjectModule
	env           map[string]string
}

// apiError, see https://www.rfc-editor.org/rfc/rfc7807
type apiError struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}
