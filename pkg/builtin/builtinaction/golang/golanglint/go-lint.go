package golanglint

import (
	_ "embed"
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/go-playground/validator/v10"
	"os"
	"path"
	"strings"
)

const URI = "builtin://actions/go-lint"

//go:embed golangci-lint-config.yml
var defaultConfig []byte

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "go-lint",
		Description: "Runs the golangci-lint tool on your go project.",
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
				{
					Name: "gitlab-sarif-converter",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
				{
					Type:   "report",
					Format: "gl-codequality",
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

	// files
	configFile := ""
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "golangci-lint.sarif.json")

	// if no config is present, create a default config
	if !a.Sdk.FileExists(path.Join(d.Module.ModuleDir, ".golangci.yml")) && !a.Sdk.FileExists(path.Join(d.Module.ModuleDir, ".golangci.yaml")) && !a.Sdk.FileExists(path.Join(d.Module.ModuleDir, ".golangci.toml")) && !a.Sdk.FileExists(path.Join(d.Module.ModuleDir, ".golangci.json")) {
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
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `golangci-lint ` + strings.Join(cmdArgs, " "),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("golangci-lint failed with exit code %d: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// store report
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	})
	if err != nil {
		return err
	}

	// optional report conversion
	err = common.GLCodeQualityConversionForModule(a.Sdk, *d, reportFile)
	if err != nil {
		return err
	}

	return nil
}
