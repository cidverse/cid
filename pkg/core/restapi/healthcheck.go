package restapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// healthCheck returns a simple up status
func (hc *APIConfig) healthCheck(c echo.Context) error {
	res := map[string]interface{}{
		"status": "up",
	}

	return c.JSON(http.StatusOK, res)
}
