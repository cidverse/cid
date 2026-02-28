package restapi

import (
	"net/http"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/labstack/echo/v5"
)

func (hc *APIConfig) healthV1(c *echo.Context) error {
	response := actionsdk.HealthV1Response{
		Status: "up",
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) logV1(c *echo.Context) error {
	var req actionsdk.LogV1Request
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "bad request",
			Details: "bad request, " + err.Error(),
		})
	}

	// log
	err = hc.SDKClient.LogV1(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "internal server error",
			Details: "failed to log message, " + err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, nil)
}

func (hc *APIConfig) uuidV4(c *echo.Context) error {
	uuid := hc.SDKClient.UUIDV4()

	return c.JSON(http.StatusOK, uuid)
}
