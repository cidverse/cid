package ansibledeploy

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestAnsibleDeploy(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExists", "/my-project/playbook-a/roles/requirements.yml").Return(false)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `ansible-playbook "/my-project/playbook-a/playbook.yml" -i "/my-project/playbook-a/inventory"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
