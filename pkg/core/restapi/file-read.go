package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// fileRead retrieves the content of a file (omitting secrets)
func (hc handlerConfig) fileRead(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}
