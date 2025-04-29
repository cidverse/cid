package helmpublishnexus

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmcommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestHelmPublishNexus(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(helmcommon.GetHelmTestData(map[string]string{
		"HELM_NEXUS_URL":        "https://localhost:9999",
		"HELM_NEXUS_REPOSITORY": "dummy",
		"HELM_NEXUS_USERNAME":   "admin",
		"HELM_NEXUS_PASSWORD":   "admin",
	}, false), nil)
	sdk.On("ArtifactList", cidsdk.ArtifactListRequest{Query: `artifact_type == "helm-chart" && format == "tgz"`}).Return(&[]cidsdk.ActionArtifact{
		{
			ID:     "root/helm-chart/mychart.tgz",
			Module: "root",
			Type:   "helm-chart",
			Name:   "mychart.tgz",
			Format: "tgz",
		},
	}, nil)
	sdk.On("ArtifactDownload", cidsdk.ArtifactDownloadRequest{
		ID:         "root/helm-chart/mychart.tgz",
		TargetFile: ".tmp/mychart.tgz",
	}).Return(nil)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://localhost:9999/service/rest/v1/components?repository=dummy", httpmock.NewStringResponder(200, ``))

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
