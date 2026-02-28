package helmbuild

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/stretchr/testify/assert"
)

func TestHelmBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(helmcommon.GetHelmTestData(map[string]string{}, false), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "helm dependency build .",
		WorkDir: "/my-project/charts/mychart",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileReadV1", "/my-project/charts/mychart/Chart.yaml").Return("name: mychart\nversion: 1.1.0", nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: "helm package . --version 1.1.0 --destination .tmp",
		WorkDir: "/my-project/charts/mychart",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-package",
		File:   ".tmp/mychart-1.1.0.tgz",
		Type:   "helm-chart",
		Format: "tgz",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
