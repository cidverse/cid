package helmpublishnexus

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestHelmPublishNexus(t *testing.T) {
	sdk := common.TestSetup(t)
	data := helmcommon.GetHelmTestData(map[string]string{
		"HELM_NEXUS_URL":        "https://localhost:9999",
		"HELM_NEXUS_REPOSITORY": "dummy",
		"HELM_NEXUS_USERNAME":   "admin",
		"HELM_NEXUS_PASSWORD":   "admin",
	}, false)
	data.Config.TempDir = t.TempDir()
	chartArchive := data.Config.TempDir + "/mychart.tgz"
	assert.NoError(t, os.WriteFile(chartArchive, []byte("chart"), 0644))
	sdk.On("ModuleExecutionContextV1").Return(data, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "helm-chart" && format == "tgz" && module == "my-package"`}).Return([]*actionsdk.Artifact{
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
		TargetFile: chartArchive,
	}).Return(nil, nil)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://localhost:9999/service/rest/v1/components?repository=dummy", httpmock.NewStringResponder(200, ``))

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
