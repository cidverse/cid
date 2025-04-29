package cargobuild

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargocommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestCargoBuildCrate(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/Cargo.toml").Return(`[package]
name = "my-crate"
version = "0.0.0"`, nil)
	sdk.On("FileExists", "src/main.rs").Return(false)
	sdk.On("FileExists", "src/lib.rs").Return(true)
	sdk.On("FileWrite", "/my-project/Cargo.toml", []byte(`[package]
name = "my-crate"
version = "2.0.0"`)).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "cargo package --allow-dirty -vv",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   "target/package/my-crate-2.0.0.crate",
		Type:   "crate",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestCargoBuildApplication(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileRead", "/my-project/Cargo.toml").Return(`[package]
name = "my-app"
version = "0.0.0"`, nil)
	sdk.On("FileExists", "src/main.rs").Return(true)
	sdk.On("FileExists", "src/lib.rs").Return(false)
	sdk.On("FileWrite", "/my-project/Cargo.toml", []byte(`[package]
name = "my-app"
version = "2.0.0"`)).Return(nil)

	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "cargo build --release -vv",
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   "target/release/my-app",
		Type:   "binary",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
