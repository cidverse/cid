package npmlint

import (
	"fmt"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/npm-lint"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "npm-lint",
		Description: "Run linting for a npm module",
		Category:    "lint",
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
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ModuleExecutionContextV1Response) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// package.json
	content, err := a.Sdk.FileReadV1(cidsdk.JoinPath(d.Module.ModuleDir, "package.json"))
	if err != nil {
		return err
	}
	pkg, err := npmcommon.ParsePackageJSON(content)
	if err != nil {
		return err
	}

	// check if script is present
	_, scriptFound := pkg.Scripts[`lint`]
	if !scriptFound {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "No lint script found in package.json"})
		return nil
	}

	// install
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `npm install`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm install failed, exit code %d", cmdResult.Code)
	}

	// lint
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `npm run lint`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm lint failed, exit code %d", cmdResult.Code)
	}

	// collect and store jacoco test reports
	testReports, err := a.Sdk.FileListV1(actionsdk.FileV1Request{
		Directory:  d.Module.ModuleDir,
		Extensions: []string{".sarif"},
	})
	for _, report := range testReports {
		_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:          report.Path,
			Module:        d.Module.Slug,
			Type:          "report",
			Format:        "sarif",
			FormatVersion: "2.1.0",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
