package restapi

import (
	"net/http"

	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v5"
)

// projectInformation returns all available information about the current project
func (hc *APIConfig) moduleList(c *echo.Context) error {
	var modules []analyzerapi.ProjectModule
	for _, module := range hc.Modules {
		module.RootDirectory = ci.ToUnixPath(module.RootDirectory)
		module.Directory = ci.ToUnixPath(module.Directory)

		var discovery []analyzerapi.ProjectModuleDiscovery
		for _, d := range module.Discovery {
			if d.File != "" {
				discovery = append(discovery, analyzerapi.ProjectModuleDiscovery{File: ci.ToUnixPath(d.File)})
			}
		}
		module.Discovery = discovery

		modules = append(modules, *module)
	}

	return c.JSON(http.StatusOK, modules)
}
