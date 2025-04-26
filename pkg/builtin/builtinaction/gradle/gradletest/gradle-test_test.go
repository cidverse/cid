package gradletest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"testing"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestGradleTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleActionDataV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION": "false",
	}, false), nil)
	sdk.On("FileExists", "/my-project/gradlew").Return(true)
	sdk.On("FileExists", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommand", cidsdk.ExecuteCommandRequest{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="1.0.0" check --no-daemon --warning-mode=all --console=plain --stacktrace`,
		WorkDir: "/my-project",
	}).Return(&cidsdk.ExecuteCommandResponse{Code: 0}, nil)
	sdk.On("FileList", cidsdk.FileRequest{Directory: "/my-project", Extensions: []string{"jacocoTestReport.xml", ".sarif", ".xml"}}).Return([]cidsdk.File{cidsdk.NewFile("/my-project/build/reports/jacoco/test/jacocoTestReport.xml")}, nil)
	sdk.On("ArtifactUpload", cidsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/build/reports/jacoco/test/jacocoTestReport.xml",
		Type:   "report",
		Format: "jacoco",
	}).Return(nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
