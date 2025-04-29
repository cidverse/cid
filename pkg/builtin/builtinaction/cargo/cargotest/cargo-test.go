package cargotest

import (
	_ "embed"
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/go-playground/validator/v10"
	"path"
)

const URI = "builtin://actions/cargo-test"

//go:embed nextest.toml
var nextestBytes []byte

type Action struct {
	Sdk cidsdk.SDKClient
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

	// create ci config
	configFile := cidsdk.JoinPath(d.Config.TempDir, "nextest.toml")
	err = a.Sdk.FileWrite(configFile, nextestBytes)
	if err != nil {
		return fmt.Errorf("error writing nextest config %s: %w", configFile, err)
	}

	// test
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`cargo nextest run --profile=ci --tool-config-file ci:%s`, configFile),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("cargo test failed, exit code %d", cmdResult.Code)
	}

	// junit report
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:   path.Join(d.Module.ModuleDir, "target", "nextest", "ci", "junit.xml"),
		Type:   "report",
		Format: "junit",
	})
	if err != nil {
		return err
	}

	return nil
}
