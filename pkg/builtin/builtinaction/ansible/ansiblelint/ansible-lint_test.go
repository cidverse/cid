package ansiblelint

import (
	_ "embed"
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/stretchr/testify/assert"
)

//go:embed report.sarif.json
var reportJson string

func TestAnsibleLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExistsV1", "/my-project/playbook-a/roles/requirements.yml").Return(false)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `ansible-lint --project . --profile "production" --sarif-file "/my-project/.tmp/ansiblelint.sarif.json"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 2}, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/ansiblelint.sarif.json").Return(reportJson, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/ansiblelint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestAnsibleLintWithDependencies(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExistsV1", "/my-project/playbook-a/roles/requirements.yml").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "ansible-galaxy install -g -f -r roles/requirements.yml -p roles",
		WorkDir: "/my-project/playbook-a",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `ansible-lint --project . --profile "production" --sarif-file "/my-project/.tmp/ansiblelint.sarif.json"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/ansiblelint.sarif.json").Return(reportJson, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/ansiblelint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
