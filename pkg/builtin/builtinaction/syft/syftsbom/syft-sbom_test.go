package syftsbom

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
)

func TestSyftSBOM(t *testing.T) {
	sdk := common.TestSetup(t)
	ctx := mavencommon.MavenTestData(map[string]string{
		"NCI_COMMIT_REF_TYPE":    "tag",
		"NCI_COMMIT_REF_RELEASE": "1.0.0",
	}, false)
	sdk.On("ModuleExecutionContextV1").Return(ctx, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `syft scan dir:. --source-name "my-module" --exclude "./.dist/**" --exclude "./.tmp/**" --source-version "1.0.0" --output "cyclonedx-json=/my-project/.tmp/syft-my-module.cdx.json" --output "spdx-json=/my-project/.tmp/syft-my-module.spdx.json"`,
		WorkDir: "/my-project",
		Env:     map[string]string{"SYFT_CHECK_FOR_APP_UPDATE": "false"},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/syft-my-module.cdx.json").Return(`{"bomFormat":"CycloneDX","specVersion":"1.6"}`, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/syft-my-module.spdx.json").Return(`{"spdxVersion":"SPDX-2.3"}`, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/syft-my-module.cdx.json",
		Module:        "my-module",
		Type:          "sbom",
		Format:        "cyclonedx-json",
		FormatVersion: "1.6",
	}).Return("", "", nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/syft-my-module.spdx.json",
		Module:        "my-module",
		Type:          "sbom",
		Format:        "spdx-json",
		FormatVersion: "2.3",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
