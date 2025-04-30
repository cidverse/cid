package cargobuild

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargocommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/lib/formats/cargotoml"
)

const URI = "builtin://actions/cargo-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	CargoVersion string `json:"cargo_version"        env:"CARGO_VERSION"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "cargo-build",
		Description: "Builds a Rust project using cargo.",
		Category:    "build",
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
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type: "binary",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}
	if cfg.CargoVersion == "" {
		cfg.CargoVersion = cargocommon.GetVersion(d.Env["NCI_COMMIT_REF_TYPE"], d.Env["NCI_COMMIT_REF_RELEASE"], d.Env["NCI_COMMIT_HASH_SHORT"])
	}

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
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// read cargo package
	cargoTomlFile := cidsdk.JoinPath(d.Module.ModuleDir, "Cargo.toml")
	cargoBytes, err := a.Sdk.FileRead(cargoTomlFile)
	if err != nil {
		return fmt.Errorf("error reading cargo.toml: %v", err)
	}

	mainExists := a.Sdk.FileExists("src/main.rs")
	libExists := a.Sdk.FileExists("src/lib.rs")

	// TD-003: patch version in Cargo.toml due to cargo limitations
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "patching cargo.toml version", Context: map[string]interface{}{"version": cfg.CargoVersion}})
	patchedCargoBytes, err := cargotoml.PatchVersion([]byte(cargoBytes), cfg.CargoVersion)
	if err != nil {
		return err
	}
	err = a.Sdk.FileWrite(cargoTomlFile, patchedCargoBytes)
	if err != nil {
		return fmt.Errorf("error writing patched cargo.toml: %v", err)
	}

	// parse cargo package
	packageConfig, err := cargotoml.ReadBytes(patchedCargoBytes)
	if err != nil {
		return fmt.Errorf("error parsing cargo.toml: %v", err)
	}

	// build (executable)
	if mainExists {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "main.rs found, building executable"})

		cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: `cargo build --release -vv`,
			WorkDir: d.Module.ModuleDir,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("cargo build failed, exit code %d", cmdResult.Code)
		}

		err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
			File:   fmt.Sprintf("target/release/%s", packageConfig.Package.Name),
			Module: d.Module.Slug,
			Type:   "binary",
		})
		if err != nil {
			return err
		}
	}

	// build (crate)
	if libExists {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "lib.rs found, building crate"})

		cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: `cargo package --allow-dirty -vv`,
			WorkDir: d.Module.ModuleDir,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("cargo build failed, exit code %d", cmdResult.Code)
		}

		err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
			File:   fmt.Sprintf("target/package/%s-%s.crate", packageConfig.Package.Name, packageConfig.Package.Version),
			Module: d.Module.Slug,
			Type:   "crate",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
