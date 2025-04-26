package semgrepscan

import (
	_ "embed"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestSemgrepScan(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectActionDataV1").Return(common.TestProjectData(), nil)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `semgrep ci --text --sarif-output="/my-project/.tmp/semgrep.sarif.json" --metrics=off --disable-version-check --exclude=.dist --exclude=.tmp --config "p/ci"`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"SEMGREP_APP_TOKEN": "",
			"SEMGREP_RULES":     "",
		},
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/semgrep.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
