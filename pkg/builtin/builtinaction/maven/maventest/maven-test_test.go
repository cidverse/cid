package maventest

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenTest(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(mavencommon.MavenTestData(map[string]string{
		"WRAPPER_VERIFICATION": "false",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/mvnw").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -classpath=".mvn/wrapper/maven-wrapper.jar" org.apache.maven.wrapper.MavenWrapperMain versions:set -DnewVersion="1.0.0"`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -classpath=".mvn/wrapper/maven-wrapper.jar" org.apache.maven.wrapper.MavenWrapperMain test --batch-mode`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("FileListV1", actionsdk.FileV1Request{Directory: "/my-project", Extensions: []string{".xml", ".sarif"}}).Return([]actionsdk.File{actionsdk.NewFile("/my-project/target/site/jacoco/jacoco.xml")}, nil)
	sdk.On("ArtifactUploadV1", actionsdk.ArtifactUploadRequest{
		Module: "my-module",
		File:   "/my-project/target/site/jacoco/jacoco.xml",
		Type:   "report",
		Format: "jacoco",
	}).Return("", "", nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
