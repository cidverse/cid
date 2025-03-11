package restapi

import (
	"net/http"

	"github.com/cidverse/cid/pkg/core/deployment"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
)

type DeploymentResponse struct {
	DeploymentType string            `json:"deployment_type"`
	DeploymentSpec string            `json:"deployment_spec"`
	DeploymentFile string            `json:"deployment_file"`
	Properties     map[string]string `json:"properties"`
}

// moduleCurrent returns information about the current module if the action is module-scoped (config)
func (hc *APIConfig) deployment(c echo.Context) error {
	var module = hc.CurrentModule
	if module == nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "module-scoped action called without module context",
			Details: "Configuration error: Deployment actions are required to be module-scoped, but no module context is available",
		})
	}
	if module.DeploymentSpec == "" {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "current module is not a deployment",
			Details: "current module is not a deployment",
		})
	}

	response := DeploymentResponse{}
	if module.DeploymentSpec == analyzerapi.DeploymentSpecDotEnv && len(module.Discovery) > 0 {
		// assign module information
		response.DeploymentType = module.DeploymentType
		response.DeploymentSpec = string(module.DeploymentSpec)
		response.DeploymentFile = module.Discovery[0].File

		// read dot env file
		properties, err := deployment.ParseDotEnvFile(response.DeploymentFile)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, apiError{
				Status:  500,
				Title:   "failed to read deployment file",
				Details: err.Error(),
			})
		}
		response.Properties = properties
	}

	return c.JSON(http.StatusOK, response)
}
