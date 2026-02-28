package gradletest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradleTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION": "false",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/gradlew").Return(true)
	sdk.On("FileExistsV1", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="1.0.0" check --no-daemon --warning-mode=all --console=plain --stacktrace`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project", Extensions: []string{"jacocoTestReport.xml", ".sarif", ".xml"}}).Return([]actionsdk.File{actionsdk.NewFile("/my-project/build/reports/jacoco/test/jacocoTestReport.xml")}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/build/reports/jacoco/test/jacocoTestReport.xml",
		Type:   "report",
		Format: "jacoco",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
