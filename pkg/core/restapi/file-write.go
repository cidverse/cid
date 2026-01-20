package restapi

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// fileWrite writes the content into the specified file, dirs and files will be created if not present
func (hc *APIConfig) fileWrite(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}
