package gitleaksscan

import (
	_ "embed"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed report.sarif.json
var reportJson string

func TestGitleaksScanBuild(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `gitleaks detect --source=. -v --no-git --report-format=sarif --report-path="/my-project/.tmp/gitleaks.sarif.json" --no-banner --redact=85 --exit-code 0`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileReadV1", "/my-project/.tmp/gitleaks.sarif.json").Return(reportJson, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		File:          "/my-project/.tmp/gitleaks.sarif.json",
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
