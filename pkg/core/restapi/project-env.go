package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// projectEnv returns the available env for this action
func (hc handlerConfig) projectEnv(c echo.Context) error {
	if hc.env == nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "env is not accessible",
			Details: "env is not accessible",
		})
	}

	return c.JSON(http.StatusOK, hc.env)
}
