package golangcilint

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/golangci-lint"

//go:embed golangci-lint-config.yml
var defaultConfig []byte

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "golangci-lint",
		Description: "Runs golangci-lint to check the code quality of your go project.",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "gomod"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "golangci-lint",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "proxy.golang.org:443",
				},
				{
					Host: "storage.googleapis.com:443",
				},
				{
					Host: "sum.golang.org:443",
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

	// files
	configFile := ""
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "golangci-lint.sarif.json")

	// if no config is present, create a default config
	if !a.Sdk.FileExistsV1(path.Join(d.Module.ModuleDir, ".golangci.yml")) && !a.Sdk.FileExistsV1(path.Join(d.Module.ModuleDir, ".golangci.yaml")) && !a.Sdk.FileExistsV1(path.Join(d.Module.ModuleDir, ".golangci.toml")) && !a.Sdk.FileExistsV1(path.Join(d.Module.ModuleDir, ".golangci.json")) {
		configFile = cidsdk.JoinPath(d.Config.TempDir, ".golangci.yml")

		err = os.WriteFile(configFile, defaultConfig, 0644)
		if err != nil {
			return err
		}
	}

	// execute
	cmdArgs := []string{
		"run",
		"--output.text.path stdout",
		fmt.Sprintf("--output.sarif.path %q", reportFile),
		"--issues-exit-code 0",
	}
	if configFile != "" {
		cmdArgs = append(cmdArgs, "--config", configFile)
	}
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `golangci-lint ` + strings.Join(cmdArgs, " "),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("golangci-lint failed with exit code %d: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// store report
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	})
	if err != nil {
		return err
	}

	return nil
}
