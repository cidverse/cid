package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
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

	return c.JSON(http.StatusOK, hc.currentModule)
}
