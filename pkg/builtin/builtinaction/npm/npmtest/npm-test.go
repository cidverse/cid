package npmtest

import (
	"fmt"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"
)

const URI = "builtin://actions/npm-test"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "npm-test",
		Description: "Run tests for a npm module",
		Category:    "test",
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
					Format: "cobertura",
				},
				{
					Type:   "report",
					Format: "junit",
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
	_, scriptFound := pkg.Scripts[`test`]
	if !scriptFound {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "No test script found in package.json"})
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

	// test
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `npm test`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm test failed, exit code %d", cmdResult.Code)
	}

	// collect and store jacoco test reports
	testReports, err := a.Sdk.FileListV1(actionsdk.FileV1Request{
		Directory:  d.Module.ModuleDir,
		Extensions: []string{".xml"},
	})
	for _, report := range testReports {
		if strings.HasSuffix(report.Path, "cobertura-coverage.xml") {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "cobertura",
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(report.Path, "junit.xml") {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "junit",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
