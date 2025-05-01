package ansiblelint

import (
	_ "embed"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

//go:embed report.sarif.json
var reportJson string

func TestAnsibleLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExists", "/my-project/playbook-a/roles/requirements.yml").Return(false)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `ansible-lint --project . --profile "production" --sarif-file "/my-project/.tmp/ansiblelint.sarif.json"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 2}, nil)
	sdk.On("FileRead", "/my-project/.tmp/ansiblelint.sarif.json").Return(reportJson, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/ansiblelint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestAnsibleLintWithDependencies(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExists", "/my-project/playbook-a/roles/requirements.yml").Return(true)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: "ansible-galaxy install -g -f -r roles/requirements.yml -p roles",
		WorkDir: "/my-project/playbook-a",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `ansible-lint --project . --profile "production" --sarif-file "/my-project/.tmp/ansiblelint.sarif.json"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("FileRead", "/my-project/.tmp/ansiblelint.sarif.json").Return(reportJson, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/ansiblelint.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
