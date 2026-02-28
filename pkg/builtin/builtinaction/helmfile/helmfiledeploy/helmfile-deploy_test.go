package helmfiledeploy

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmfileDeploy(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(helmcommon.GetHelmfileTestData(map[string]string{
		"DEPLOYMENT_NAMESPACE": "temp",
		"KUBECONFIG_BASE64":    "YXBpVmVyc2lvbjogdjEKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eS1kYXRhOiA8Y2EtZGF0YS1oZXJlPgogICAgc2VydmVyOiBodHRwczovL3lvdXItazhzLWNsdXN0ZXIuY29tCiAgbmFtZTogPGNsdXN0ZXItbmFtZT4KY29udGV4dHM6Ci0gY29udGV4dDoKICAgIGNsdXN0ZXI6ICA8Y2x1c3Rlci1uYW1lPgogICAgdXNlcjogIDxjbHVzdGVyLW5hbWUtdXNlcj4KICBuYW1lOiAgPGNsdXN0ZXItbmFtZT4KY3VycmVudC1jb250ZXh0OiAgPGNsdXN0ZXItbmFtZT4Ka2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQp1c2VyczoKLSBuYW1lOiAgPGNsdXN0ZXItbmFtZS11c2VyPgogIHVzZXI6CiAgICB0b2tlbjogPHNlY3JldC10b2tlbi1oZXJlPg==",
	}, false), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "helmfile init --force",
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `helmfile apply -f "helmfile.yaml" --namespace="temp" --environment="dev" --suppress-diff `,
		WorkDir: "/my-project",
		Env: map[string]string{
			"KUBECONFIG": ".tmp/kube/kubeconfig",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
