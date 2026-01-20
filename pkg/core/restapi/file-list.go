package restapi

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// fileList retrieves a list of all files
func (hc *APIConfig) fileList(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}
