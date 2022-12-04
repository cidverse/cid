package restapi

import (
	"net/http"

	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
)

// moduleCurrent returns information about the current module if the action is module-scoped (config)
func (hc *handlerConfig) moduleCurrent(c echo.Context) error {
	if hc.currentModule == nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "no current module when action is running in project scope",
			Details: "no current module when action is running in project scope, actions need to be module scoped for access the current module",
		})
	}

	var module = hc.currentModule
	module.RootDirectory = cihelper.ToUnixPath(module.RootDirectory)
	module.Directory = cihelper.ToUnixPath(module.Directory)

	var discovery []analyzerapi.ProjectModuleDiscovery
	for _, d := range module.Discovery {
		if d.File != "" {
			discovery = append(discovery, analyzerapi.ProjectModuleDiscovery{File: cihelper.ToUnixPath(d.File)})
		}
	}
	module.Discovery = discovery

	return c.JSON(http.StatusOK, module)
}
