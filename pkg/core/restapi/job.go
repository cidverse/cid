package restapi

import (
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/core/deployment"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
)

type JobModuleDataResponse struct {
	ProjectDir string                 `json:"project-dir"`
	Config     map[string]interface{} `json:"config"`
	Env        map[string]string      `json:"env"`
	Module     *analyzerapi.ProjectModule
	Deployment *DeploymentResponse
}

// jobModuleDataV1 returns data for module-scoped actions
func (hc *APIConfig) jobModuleDataV1(c echo.Context) error {
	response := JobModuleDataResponse{
		ProjectDir: hc.ProjectDir,
		Config:     JobConfigDataV1(hc.ProjectDir, hc.TempDir, hc.ActionConfig),
		Env:        hc.ActionEnv,
		Module:     hc.CurrentModule,
		Deployment: nil,
	}

	if hc.CurrentModule.DeploymentSpec != "" {
		resp, err := JobDeploymentV1(hc.CurrentModule, hc.ActionEnv, hc.Env)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, apiError{
				Status:  500,
				Title:   "error reading deployment file",
				Details: err.Error(),
			})
		}

		// merge deployment properties into env, env takes precedence
		for k, v := range resp.Properties {
			if _, exists := response.Env[k]; !exists {
				response.Env[k] = v
			}
		}

		response.Deployment = &resp
	}

	return c.JSON(http.StatusOK, response)
}

type JobProjectDataResponse struct {
	ProjectDir string                 `json:"project-dir"`
	Config     map[string]interface{} `json:"config"`
	Env        map[string]string      `json:"env"`
}

// jobProjectDataV1 returns data for project-scoped actions
func (hc *APIConfig) jobProjectDataV1(c echo.Context) error {
	response := JobProjectDataResponse{
		ProjectDir: hc.ProjectDir,
		Config:     JobConfigDataV1(hc.ProjectDir, hc.TempDir, hc.ActionConfig),
		Env:        hc.ActionEnv,
	}
	return c.JSON(http.StatusOK, response)
}

// jobConfigV1 returns the configuration for the current action
func (hc *APIConfig) jobConfigV1(c echo.Context) error {
	response := JobConfigDataV1(hc.ProjectDir, hc.TempDir, hc.ActionConfig)
	return c.JSON(http.StatusOK, response)
}

func JobConfigDataV1(projectDir string, tempDir string, actionConfig string) map[string]interface{} {
	host, _ := os.Hostname()
	currentUser, _ := user.Current()

	result := map[string]interface{}{
		// enable debugging
		"debug": false,
		// toggle debug output for specific parts of the process
		"log": map[string]string{
			"bin-helm": "debug",
		},
		// host
		"host_name":      host,
		"host_user_id":   currentUser.Uid,
		"host_user_name": currentUser.Username,
		"host_group_id":  currentUser.Gid,
		// paths
		"project_dir":  ci.ToUnixPath(projectDir),
		"artifact_dir": ci.ToUnixPath(filepath.Join(projectDir, ".dist")),
		"temp_dir":     tempDir,
		// dynamic config
		"config": actionConfig,
	}

	return result
}

// jobEnvV1 returns the available env for this action
func (hc *APIConfig) jobEnvV1(c echo.Context) error {
	return c.JSON(http.StatusOK, hc.ActionEnv)
}

type DeploymentResponse struct {
	DeploymentType        string            `json:"deployment_type"`
	DeploymentSpec        string            `json:"deployment_spec"`
	DeploymentFile        string            `json:"deployment_file"`
	DeploymentEnvironment string            `json:"deployment_environment"`
	Properties            map[string]string `json:"properties"`
}

// moduleCurrent returns information about the current module if the action is module-scoped (config)
func (hc *APIConfig) jobDeploymentV1(c echo.Context) error {
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
			Title:   "current module is not a jobDeploymentV1",
			Details: "current module is not a jobDeploymentV1",
		})
	}

	resp, err := JobDeploymentV1(module, hc.ActionEnv, hc.Env)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "error reading deployment file",
			Details: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func JobDeploymentV1(module *analyzerapi.ProjectModule, actionEnv map[string]string, env map[string]string) (DeploymentResponse, error) {
	decoderConfig := deployment.DecodeSecretsConfig{
		PGPPrivateKey:         env["CID_SECRET_PGP_PRIVATE_KEY"],
		PGPPrivateKeyPassword: env["CID_SECRET_PGP_PRIVATE_KEY_PASSWORD"],
	}

	if module.DeploymentSpec == analyzerapi.DeploymentSpecDotEnv && len(module.Discovery) > 0 {
		deploymentFile := module.Discovery[0].File
		response := DeploymentResponse{
			DeploymentSpec:        string(module.DeploymentSpec),
			DeploymentType:        module.DeploymentType,
			DeploymentEnvironment: module.DeploymentEnvironment,
			DeploymentFile:        deploymentFile,
			Properties:            make(map[string]string),
		}

		// read common and module specific env files
		envFiles := []string{path.Join(module.Directory, ".env-common"), deploymentFile}
		for _, f := range envFiles {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				continue
			}

			// read dot env file
			properties, err := deployment.ParseDotEnvFile(f)
			if err != nil {
				return response, err
			}

			// handle placeholders
			for k, v := range properties {
				response.Properties[k] = util.ResolveEnvPlaceholders(v, actionEnv)
			}
		}

		// decode secrets
		var err error
		response.Properties, err = deployment.DecodeSecrets(response.Properties, decoderConfig)
		if err != nil {
			return DeploymentResponse{}, err
		}

		return response, nil
	} else if module.DeploymentSpec == analyzerapi.DeploymentSpecHelmfile && len(module.Discovery) > 0 {
		deploymentFile := module.Discovery[0].File
		response := DeploymentResponse{
			DeploymentSpec:        string(module.DeploymentSpec),
			DeploymentType:        module.DeploymentType,
			DeploymentEnvironment: module.DeploymentEnvironment,
			DeploymentFile:        deploymentFile,
			Properties:            make(map[string]string),
		}

		// decode secrets
		var err error
		response.Properties, err = deployment.DecodeSecrets(response.Properties, decoderConfig)
		if err != nil {
			return DeploymentResponse{}, err
		}

		return response, nil
	}

	return DeploymentResponse{}, nil
}
