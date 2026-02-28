package restapi

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (hc *APIConfig) jobModuleDataV1(c *echo.Context) error {
	response, err := hc.SDKClient.ModuleExecutionContextV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get module execution context", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) jobProjectDataV1(c *echo.Context) error {
	response, err := hc.SDKClient.ProjectExecutionContextV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get project execution context", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) jobConfigV1(c *echo.Context) error {
	response, err := hc.SDKClient.ConfigV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get config", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) jobEnvV1(c *echo.Context) error {
	response, err := hc.SDKClient.EnvironmentV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get environment", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) jobDeploymentV1(c *echo.Context) error {
	response, err := hc.SDKClient.DeploymentV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get deployment", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) moduleList(c *echo.Context) error {
	response, err := hc.SDKClient.ModuleListV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get module list", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) moduleCurrent(c *echo.Context) error {
	response, err := hc.SDKClient.ModuleCurrentV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get current module", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}
