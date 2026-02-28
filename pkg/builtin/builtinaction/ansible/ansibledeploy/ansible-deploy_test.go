package ansibledeploy

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/stretchr/testify/assert"
)

func TestAnsibleDeploy(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(ansiblecommon.ModuleTestData(), nil)
	sdk.On("FileExistsV1", "/my-project/playbook-a/roles/requirements.yml").Return(false)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `ansible-playbook "/my-project/playbook-a/playbook.yml" -i "/my-project/playbook-a/inventory"`,
		WorkDir: "/my-project/playbook-a",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
