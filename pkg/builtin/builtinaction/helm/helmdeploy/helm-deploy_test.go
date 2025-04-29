package helmdeploy

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestHelmDeploy(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(helmcommon.GetHelmTestData(map[string]string{
		"DEPLOYMENT_CHART":         "oci://registry-1.docker.io/bitnamicharts/nginx",
		"DEPLOYMENT_CHART_VERSION": "19.0.1",
		"DEPLOYMENT_NAMESPACE":     "temp",
		"DEPLOYMENT_ENVIRONMENT":   "stage",
		"DEPLOYMENT_ID":            "test-deployment",
		"KUBECONFIG_BASE64":        "YXBpVmVyc2lvbjogdjEKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eS1kYXRhOiA8Y2EtZGF0YS1oZXJlPgogICAgc2VydmVyOiBodHRwczovL3lvdXItazhzLWNsdXN0ZXIuY29tCiAgbmFtZTogPGNsdXN0ZXItbmFtZT4KY29udGV4dHM6Ci0gY29udGV4dDoKICAgIGNsdXN0ZXI6ICA8Y2x1c3Rlci1uYW1lPgogICAgdXNlcjogIDxjbHVzdGVyLW5hbWUtdXNlcj4KICBuYW1lOiAgPGNsdXN0ZXItbmFtZT4KY3VycmVudC1jb250ZXh0OiAgPGNsdXN0ZXItbmFtZT4Ka2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQp1c2VyczoKLSBuYW1lOiAgPGNsdXN0ZXItbmFtZS11c2VyPgogIHVzZXI6CiAgICB0b2tlbjogPHNlY3JldC10b2tlbi1oZXJlPg==",
	}, false), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command:       `helm show chart --version "19.0.1" "oci://registry-1.docker.io/bitnamicharts/nginx"`,
		WorkDir:       "/my-project/charts/mychart",
		CaptureOutput: true,
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `helm pull --untar --destination ".tmp/helm-charts" --version "19.0.1" "oci://registry-1.docker.io/bitnamicharts/nginx"`,
		WorkDir: "/my-project/charts/mychart",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `helm upgrade --namespace "temp" --install --disable-openapi-validation  "test-deployment" ".tmp/helm-charts"`,
		WorkDir: "/my-project/charts/mychart",
		Env: map[string]string{
			"KUBECONFIG": ".tmp/kube/kubeconfig",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
