package nodebuild

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/node/nodecommon"
	"github.com/go-playground/validator/v10"
)

const URI = "builtin://actions/node-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "node-build",
		Description: "Builds a node.js project",
		Category:    "build",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "npm"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "npm",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "registry.npmjs.org:443",
				},
				{
					Host: "registry.yarnpkg.com:443",
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

	// package.json
	content, err := a.Sdk.FileRead(cidsdk.JoinPath(d.Module.ModuleDir, "package.json"))
	if err != nil {
		return err
	}
	pkg, err := nodecommon.ParsePackageJSON(content)
	if err != nil {
		return err
	}

	// check if script is present
	_, scriptFound := pkg.Scripts[`build`]
	if !scriptFound {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "warn", Message: "No build script found in package.json"})
		return nil
	}

	// install
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `npm install`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm install failed, exit code %d", cmdResult.Code)
	}

	// build
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `npm build`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm build failed, exit code %d", cmdResult.Code)
	}

	return nil
}
