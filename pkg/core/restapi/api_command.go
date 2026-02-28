package restapi

import (
	"fmt"
	"net/http"

	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/labstack/echo/v5"
)

// commandExecute runs a command in the project directory (blocking until the command exits, returns the response code)
func (hc *APIConfig) commandExecute(c *echo.Context) error {
	var req actionsdk.ExecuteCommandV1Request
	err := c.Bind(&req)
	if err != nil {
		return hc.handleError(c, http.StatusBadRequest, "invalid request body", fmt.Sprintf("invalid request body: %v", err))
	}

	resp, err := hc.SDKClient.ExecuteCommandV1(req)
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "command execution failed", fmt.Sprintf("command execution failed: %v", err))
	}

	return c.JSON(http.StatusOK, resp)
}
