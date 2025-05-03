package dotnettest

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"strings"
)

const URI = "builtin://actions/dotnet-test"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "dotnet-test",
		Description: `Runs the dotnet test command`,
		Category:    "test",
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
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "junit",
				},
				{
					Type:   "report",
					Format: "cobertura",
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

	// files
	junitReport := cidsdk.JoinPath(d.Config.TempDir, "junit.xml")
	trxReport := cidsdk.JoinPath(d.Config.TempDir, "vstest.trx")

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

	// test
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`dotnet test --logger:"junit;LogFilePath=%s;MethodFormat=Class;FailureBodyFormat=Verbose" --logger:"trx;LogFileName=%s" --collect "Code Coverage;Format=cobertura"`, junitReport, trxReport),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("dotnet test failed, exit code %d", cmdResult.Code)
	}

	// store report
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:   trxReport,
		Module: d.Module.Slug,
		Type:   "report",
		Format: "trx",
	})
	if err != nil {
		return err
	}
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:   junitReport,
		Module: d.Module.Slug,
		Type:   "report",
		Format: "junit",
	})
	if err != nil {
		return err
	}

	// collect and store cobertura reports
	testReports, err := a.Sdk.FileList(cidsdk.FileRequest{
		Directory:  d.Module.ModuleDir,
		Extensions: []string{".xml"},
	})
	for _, report := range testReports {
		if strings.HasSuffix(report.Path, ".cobertura.xml") {
			err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "cobertura",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
