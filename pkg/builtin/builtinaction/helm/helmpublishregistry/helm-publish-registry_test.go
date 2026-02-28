package helmpublishregistry

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmPublishRegistry(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(helmcommon.GetHelmTestData(map[string]string{
		"HELM_OCI_REPOSITORY": "localhost:5000/helm-charts",
	}, false), nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "helm-chart" && format == "tgz"`}).Return([]*actionsdk.Artifact{
		{
			ArtifactID: "root/helm-chart/mychart.tgz",
			Module:     "root",
			Type:       "helm-chart",
			Name:       "mychart.tgz",
			Format:     "tgz",
		},
	}, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "root/helm-chart/mychart.tgz",
		TargetFile: ".tmp/mychart.tgz",
	}).Return(nil, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `helm push .tmp/mychart.tgz oci://localhost:5000/helm-charts`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
