package restapi

import (
	"net/http"

	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
)

// projectInformation returns all available information about the current project
func (hc *handlerConfig) moduleList(c echo.Context) error {
	var modules []analyzerapi.ProjectModule
	for _, module := range hc.modules {
		module.RootDirectory = cihelper.ToUnixPath(module.RootDirectory)
		module.Directory = cihelper.ToUnixPath(module.Directory)

		var discovery []analyzerapi.ProjectModuleDiscovery
		for _, d := range module.Discovery {
			if d.File != "" {
				discovery = append(discovery, analyzerapi.ProjectModuleDiscovery{File: cihelper.ToUnixPath(d.File)})
			}
		}
		module.Discovery = discovery

		modules = append(modules, *module)
	}

	return c.JSON(http.StatusOK, modules)
}
