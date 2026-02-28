package restapi

import (
	"github.com/labstack/echo/v5"
)

func (hc *APIConfig) handleError(c *echo.Context, status int, title string, details string) error {
	return c.JSON(status, apiError{
		Status:  status,
		Title:   title,
		Details: details,
	})
}
