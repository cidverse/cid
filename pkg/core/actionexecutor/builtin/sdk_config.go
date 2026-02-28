package builtin

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/deployment"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/core/ci"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

func (sdk ActionSDK) ConfigV1() (*actionsdk.ConfigV1Response, error) {
	host, _ := os.Hostname()
	currentUser := util.GetCurrentUser()

	return &actionsdk.ConfigV1Response{
		Debug: false,
		Log: map[string]string{
			"bin-helm": "debug",
		},
		ProjectDir:   ci.ToUnixPath(sdk.ProjectDir),
		ArtifactDir:  ci.ToUnixPath(filepath.Join(sdk.ProjectDir, ".dist")),
		TempDir:      sdk.TempDir,
		HostName:     host,
		HostUserId:   currentUser.Uid,
		HostUserName: currentUser.Username,
		HostGroupId:  currentUser.Gid,
		Config:       sdk.ActionConfig,
	}, nil
}

func (sdk ActionSDK) ProjectExecutionContextV1() (*actionsdk.ProjectExecutionContextV1Response, error) {
	cfg, err := sdk.ConfigV1()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error getting config for project execution context"), err)
	}

	modules := make([]*actionsdk.ProjectModule, len(sdk.Modules))
	for i, module := range sdk.Modules {
		modules[i] = convertProjectModule(module)
	}

	response := actionsdk.ProjectExecutionContextV1Response{
		ProjectDir: sdk.ProjectDir,
		Config:     cfg,
		Env:        sdk.ActionEnv,
		Modules:    modules,
	}

	return &response, nil
}

func (sdk ActionSDK) ModuleExecutionContextV1() (*actionsdk.ModuleExecutionContextV1Response, error) {
	cfg, err := sdk.ConfigV1()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error getting config for project execution context"), err)
	}

	response := actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: sdk.ProjectDir,
		Config:     cfg,
		Env:        sdk.ActionEnv,
		Module:     convertProjectModule(sdk.CurrentModule),
		Deployment: nil,
	}

	if sdk.CurrentModule.DeploymentSpec != "" {
		resp, err := sdk.DeploymentV1()
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error reading deployment information for current module"), err)
		}

		// merge deployment properties into env, env takes precedence
		for k, v := range resp.Properties {
			if _, exists := response.Env[k]; !exists {
				response.Env[k] = v
			}
		}

		response.Deployment = resp
	}

	return &response, nil
}

func (sdk ActionSDK) EnvironmentV1() (*actionsdk.EnvironmentV1Response, error) {
	return &actionsdk.EnvironmentV1Response{
		Env: sdk.ActionEnv,
	}, nil
}

func (sdk ActionSDK) DeploymentV1() (*actionsdk.DeploymentV1Response, error) {
	module := sdk.CurrentModule
	if module == nil {
		return nil, fmt.Errorf("no current module when action is running in project scope, actions need to be module scoped for access the current module")
	}
	if module.DeploymentSpec == "" {
		return nil, fmt.Errorf("current module is not a deployment, deployment information is only available for modules with a deployment specification")
	}

	decoderConfig := deployment.DecodeSecretsConfig{
		PGPPrivateKey:         sdk.Env["CID_SECRET_PGP_PRIVATE_KEY"],
		PGPPrivateKeyPassword: sdk.Env["CID_SECRET_PGP_PRIVATE_KEY_PASSWORD"],
	}

	if module.DeploymentSpec == analyzerapi.DeploymentSpecDotEnv && len(module.Discovery) > 0 {
		deploymentFile := module.Discovery[0].File
		response := actionsdk.DeploymentV1Response{
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
				return &response, err
			}

			// handle placeholders
			for k, v := range properties {
				response.Properties[k] = util.ResolveEnvPlaceholders(v, sdk.ActionEnv)
			}
		}

		// decode secrets
		var err error
		response.Properties, err = deployment.DecodeSecrets(response.Properties, decoderConfig)
		if err != nil {
			return nil, err
		}

		return &response, nil
	} else if module.DeploymentSpec == analyzerapi.DeploymentSpecHelmfile && len(module.Discovery) > 0 {
		deploymentFile := module.Discovery[0].File
		response := actionsdk.DeploymentV1Response{
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
			return nil, err
		}

		return &response, nil
	}

	return nil, nil
}

func (sdk ActionSDK) ModuleListV1() ([]*actionsdk.ProjectModule, error) {
	var modules []*actionsdk.ProjectModule
	for _, module := range sdk.Modules {
		module.RootDirectory = ci.ToUnixPath(module.RootDirectory)
		module.Directory = ci.ToUnixPath(module.Directory)

		var discovery []analyzerapi.ProjectModuleDiscovery
		for _, d := range module.Discovery {
			if d.File != "" {
				discovery = append(discovery, analyzerapi.ProjectModuleDiscovery{File: ci.ToUnixPath(d.File)})
			}
		}
		module.Discovery = discovery

		modules = append(modules, convertProjectModule(module))
	}

	return modules, nil
}

func (sdk ActionSDK) ModuleCurrentV1() (*actionsdk.ProjectModule, error) {
	if sdk.CurrentModule == nil {
		return nil, fmt.Errorf("no current module when action is running in project scope, actions need to be module scoped for access the current module")
	}

	var module = sdk.CurrentModule
	return convertProjectModule(module), nil
}

func convertProjectModule(module *analyzerapi.ProjectModule) *actionsdk.ProjectModule {
	var discovery []actionsdk.ProjectModuleDiscovery
	for _, d := range module.Discovery {
		if d.File != "" {
			discovery = append(discovery, actionsdk.ProjectModuleDiscovery{File: ci.ToUnixPath(d.File)})
		}
	}

	var files = make([]string, len(module.Files))
	for _, file := range module.Files {
		files = append(files, ci.ToUnixPath(file))
	}

	// submodules
	var submodules []*actionsdk.ProjectModule
	for _, submodule := range module.Submodules {
		submodules = append(submodules, convertProjectModule(submodule))
	}

	// languages
	languages := make(map[string]string)
	for k, v := range module.Language {
		languages[string(k)] = v
	}

	// dependencies
	var dependencies []*actionsdk.ProjectDependency
	for _, dep := range module.Dependencies {
		dependencies = append(dependencies, &actionsdk.ProjectDependency{
			Id:      dep.ID,
			Type:    dep.Type,
			Version: dep.Version,
			Scope:   dep.Scope,
		})
	}

	return &actionsdk.ProjectModule{
		ProjectDir:            ci.ToUnixPath(module.RootDirectory),
		ModuleDir:             ci.ToUnixPath(module.Directory),
		Discovery:             discovery,
		Name:                  module.Name,
		Slug:                  module.Slug,
		Type:                  string(module.Type),
		BuildSystem:           string(module.BuildSystem),
		BuildSystemSyntax:     string(module.BuildSystemSyntax),
		SpecificationType:     string(module.SpecificationType),
		ConfigType:            string(module.ConfigType),
		DeploymentSpec:        string(module.DeploymentSpec),
		DeploymentType:        module.DeploymentType,
		DeploymentEnvironment: module.DeploymentEnvironment,
		Language:              languages,
		Dependencies:          dependencies,
		Files:                 files,
		Submodules:            submodules,
	}
}
