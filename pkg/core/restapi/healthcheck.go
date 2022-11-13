package restapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// healthCheck returns a simple up status
func (hc handlerConfig) healthCheck(c echo.Context) error {
	res := map[string]interface{}{
		"status": "up",
	}

	return c.JSON(http.StatusOK, res)
}
