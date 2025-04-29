package dotnetbuild

import (
	"fmt"
	"github.com/go-playground/validator/v10"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/dotnet-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "dotnet-build",
		Description: `Builds a .NET project using dotnet.`,
		Category:    "build",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "dotnet"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "dotnet",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "api.nuget.org:443",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// validate
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// restore
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`dotnet restore`),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("dotnet restore failed, exit code %d", cmdResult.Code)
	}

	// build
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`dotnet build --runtime linux-x64 --self-contained --configuration Release`),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("dotnet build failed, exit code %d", cmdResult.Code)
	}

	return nil
}
