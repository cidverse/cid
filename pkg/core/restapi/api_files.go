package restapi

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// fileList retrieves a list of all files
func (hc *APIConfig) fileList(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}

// fileRead retrieves the content of a file (omitting secrets)
func (hc *APIConfig) fileRead(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}

// fileWrite writes the content into the specified file, dirs and files will be created if not present
func (hc *APIConfig) fileWrite(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, apiError{
		Status:  500,
		Title:   "not yet implemented",
		Details: "not yet implemented",
	})
}
