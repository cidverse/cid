package helmlint

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/stretchr/testify/assert"
)

func TestHelmLint(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(helmcommon.GetHelmTestData(map[string]string{}, false), nil)
	sdk.On("FileReadV1", "/my-project/charts/mychart/Chart.yaml").Return("name: mychart\nversion: 1.1.0", nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "helm lint . --strict",
		WorkDir: "/my-project/charts/mychart",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
