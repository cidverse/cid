package trivyfsscan

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrivyFSScan(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `trivy fs . --severity MEDIUM,HIGH,CRITICAL --format sarif --output /my-project/.tmp/trivyfs.sarif.json`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{
		Code: 0,
	}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/trivyfs.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
