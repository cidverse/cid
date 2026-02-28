package cargotest

import (
	_ "embed"
	"fmt"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"path"
)

const URI = "builtin://actions/cargo-test"

//go:embed nextest.toml
var nextestBytes []byte

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "cargo-test",
		Description: "Tests a Rust project",
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "cargo"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "cargo",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "crates.io:443",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
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

	// create ci config
	configFile := cidsdk.JoinPath(d.Config.TempDir, "nextest.toml")
	err = a.Sdk.FileWriteV1(configFile, nextestBytes)
	if err != nil {
		return fmt.Errorf("error writing nextest config %s: %w", configFile, err)
	}

	// test
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: fmt.Sprintf(`cargo nextest run --profile=ci --tool-config-file ci:%s`, configFile),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("cargo test failed, exit code %d", cmdResult.Code)
	}

	// junit report
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:   path.Join(d.Module.ModuleDir, "target", "nextest", "ci", "junit.xml"),
		Type:   "report",
		Format: "junit",
	})
	if err != nil {
		return err
	}

	return nil
}
