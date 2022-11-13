package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// projectInformation returns all available information about the current project
func (hc handlerConfig) moduleList(c echo.Context) error {
	return c.JSON(http.StatusOK, hc.modules)
}
