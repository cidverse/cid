package cargobuild

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargocommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCargoBuildCrate(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileReadV1", "/my-project/Cargo.toml").Return(`[package]
name = "my-crate"
version = "0.0.0"`, nil)
	sdk.On("FileExistsV1", "src/main.rs").Return(false)
	sdk.On("FileExistsV1", "src/lib.rs").Return(true)
	sdk.On("FileWriteV1", "/my-project/Cargo.toml", []byte(`[package]
name = "my-crate"
version = "2.0.0"`)).Return(nil)

	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "cargo package --allow-dirty -vv",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   "target/package/my-crate-2.0.0.crate",
		Type:   "crate",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestCargoBuildApplication(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(cargocommon.TestModuleData(), nil)
	sdk.On("FileReadV1", "/my-project/Cargo.toml").Return(`[package]
name = "my-app"
version = "0.0.0"`, nil)
	sdk.On("FileExistsV1", "src/main.rs").Return(true)
	sdk.On("FileExistsV1", "src/lib.rs").Return(false)
	sdk.On("FileWriteV1", "/my-project/Cargo.toml", []byte(`[package]
name = "my-app"
version = "2.0.0"`)).Return(nil)

	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "cargo build --release -vv",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   "target/release/my-app",
		Type:   "binary",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
