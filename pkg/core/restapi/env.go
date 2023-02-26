package restapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// projectEnv returns the available env for this action
func (hc *APIConfig) projectEnv(c echo.Context) error {
	if hc.Env == nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "env is not accessible",
			Details: "env is not accessible",
		})
	}

	return c.JSON(http.StatusOK, hc.Env)
}
