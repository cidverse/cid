package restapi

import (
	"net/http"

	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
)

// moduleCurrent returns information about the current module if the action is module-scoped (config)
func (hc *APIConfig) moduleCurrent(c echo.Context) error {
	if hc.CurrentModule == nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "no current module when action is running in project scope",
			Details: "no current module when action is running in project scope, actions need to be module scoped for access the current module",
		})
	}

	var module = hc.CurrentModule
	module.RootDirectory = ci.ToUnixPath(module.RootDirectory)
	module.Directory = ci.ToUnixPath(module.Directory)

	var discovery []analyzerapi.ProjectModuleDiscovery
	for _, d := range module.Discovery {
		if d.File != "" {
			discovery = append(discovery, analyzerapi.ProjectModuleDiscovery{File: ci.ToUnixPath(d.File)})
		}
	}
	module.Discovery = discovery

	var files = make([]string, len(module.Files))
	for _, file := range module.Files {
		files = append(files, ci.ToUnixPath(file))
	}
	module.Files = files

	return c.JSON(http.StatusOK, module)
}
