package npmtest

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmcommon"
	"strings"
)

const URI = "builtin://actions/npm-test"

type Action struct {
	Sdk cidsdk.SDKClient
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

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
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
	pkg, err := npmcommon.ParsePackageJSON(content)
	if err != nil {
		return err
	}

	// check if script is present
	_, scriptFound := pkg.Scripts[`test`]
	if !scriptFound {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "warn", Message: "No test script found in package.json"})
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

	// test
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `npm test`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm test failed, exit code %d", cmdResult.Code)
	}

	// collect and store jacoco test reports
	testReports, err := a.Sdk.FileList(cidsdk.FileRequest{
		Directory:  d.Module.ModuleDir,
		Extensions: []string{".xml"},
	})
	for _, report := range testReports {
		if strings.HasSuffix(report.Path, "cobertura-coverage.xml") {
			err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "cobertura",
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(report.Path, "junit.xml") {
			err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
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
