package dotnettest

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"
)

const URI = "builtin://actions/dotnet-test"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "dotnet-test",
		Description: `Runs the dotnet test command`,
		Category:    "test",
		Scope:       actionsdk.ActionScopeModule,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "dotnet"`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name: "dotnet",
				},
			},
			Network: dotnetcommon.NetworkAccess,
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
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

	// all reports are written into a dedicated results directory, to support multiple test projects
	resultsDir := actionsdk.JoinPath(d.Config.TempDir, "dotnet-test")

	// restore
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `dotnet restore`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("dotnet restore failed, exit code %d", cmdResult.Code)
	}

	// test
	var testArgs []string
	if a.usesTestingPlatform(d.Module) {
		// Microsoft.Testing.Platform (opt-in via global.json since .NET 10) uses reporters instead of VSTest loggers, each reporter requires its extension package
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "Microsoft.Testing.Platform runner configured in global.json"})
		testArgs = append(testArgs, `--no-restore`, fmt.Sprintf(`--results-directory %q`, resultsDir))
		if a.referencesPackage(d.Module, "Microsoft.Testing.Extensions.TrxReport", "MSTest.Sdk") {
			testArgs = append(testArgs, `--report-trx`)
		}
		if a.referencesPackage(d.Module, "Microsoft.Testing.Extensions.CodeCoverage", "MSTest.Sdk") {
			testArgs = append(testArgs, `--coverage`, `--coverage-output-format cobertura`)
		}
		if a.referencesPackage(d.Module, "JunitXml.TestLogger") {
			testArgs = append(testArgs, fmt.Sprintf(`--report-spekt-junit "LogFilePath=%s;MethodFormat=Class;FailureBodyFormat=Verbose"`, actionsdk.JoinPath(resultsDir, "{assembly}.junit.xml")))
		}
	} else {
		testArgs = append(testArgs,
			`--no-restore`,
			fmt.Sprintf(`--results-directory %q`, resultsDir),
			`--logger:"trx"`,
			`--collect "Code Coverage;Format=cobertura"`,
		)
		if a.referencesPackage(d.Module, "JunitXml.TestLogger") {
			testArgs = append(testArgs, fmt.Sprintf(`--logger:"junit;LogFilePath=%s;MethodFormat=Class;FailureBodyFormat=Verbose"`, actionsdk.JoinPath(resultsDir, "{assembly}.junit.xml")))
		} else {
			_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "JunitXml.TestLogger package not referenced, skipping junit report"})
		}
	}
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `dotnet test ` + strings.Join(testArgs, " "),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("dotnet test failed, exit code %d", cmdResult.Code)
	}

	// collect and store reports
	testReports, err := a.Sdk.FileListV1(actionsdk.FileV1Request{
		Directory:  resultsDir,
		Extensions: []string{".trx", ".xml"},
	})
	if err != nil {
		return err
	}
	for _, report := range testReports {
		format := ""
		if strings.HasSuffix(report.Path, ".trx") {
			format = "trx"
		} else if strings.HasSuffix(report.Path, ".junit.xml") {
			format = "junit"
		} else if strings.HasSuffix(report.Path, ".cobertura.xml") {
			format = "cobertura"
		} else {
			continue
		}

		_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:   report.Path,
			Module: d.Module.Slug,
			Type:   "report",
			Format: format,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// usesTestingPlatform checks if the nearest global.json (module dir up to the project root) selects the Microsoft.Testing.Platform test runner
func (a Action) usesTestingPlatform(module *actionsdk.ProjectModule) bool {
	dir := module.ModuleDir
	for {
		globalJSON := actionsdk.JoinPath(dir, "global.json")
		if a.Sdk.FileExistsV1(globalJSON) {
			content, err := a.Sdk.FileReadV1(globalJSON)
			if err != nil {
				return false
			}

			var cfg struct {
				Test struct {
					Runner string `json:"runner"`
				} `json:"test"`
			}
			if err = json.Unmarshal([]byte(content), &cfg); err != nil {
				return false
			}

			return cfg.Test.Runner == "Microsoft.Testing.Platform"
		}

		parent := filepath.Dir(dir)
		if dir == module.ProjectDir || parent == dir || !strings.HasPrefix(dir, module.ProjectDir) {
			return false
		}
		dir = parent
	}
}

// referencesPackage checks if any project file references one of the given packages
func (a Action) referencesPackage(module *actionsdk.ProjectModule, packages ...string) bool {
	for _, file := range module.Files {
		if !strings.HasSuffix(file, "proj") && !strings.HasSuffix(file, ".props") {
			continue
		}

		content, err := a.Sdk.FileReadV1(file)
		if err != nil {
			continue
		}
		for _, pkg := range packages {
			if strings.Contains(content, pkg) {
				return true
			}
		}
	}

	return false
}
