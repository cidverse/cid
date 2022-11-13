package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// fileList retrieves a list of all files
func (hc handlerConfig) fileList(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}
