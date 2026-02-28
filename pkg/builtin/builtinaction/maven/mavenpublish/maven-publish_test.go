package mavenpublish

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMavenPublish(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(mavencommon.MavenTestData(map[string]string{
		"WRAPPER_VERIFICATION": "false",
		"MAVEN_REPO_URL":       "http://localhost:9100/test",
		"MAVEN_REPO_USERNAME":  "admin",
		"MAVEN_REPO_PASSWORD":  "secret",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/mvnw").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -classpath=".mvn/wrapper/maven-wrapper.jar" org.apache.maven.wrapper.MavenWrapperMain versions:set -DnewVersion="1.0.0"`,
		WorkDir: "/my-project",
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -classpath=".mvn/wrapper/maven-wrapper.jar" org.apache.maven.wrapper.MavenWrapperMain deploy --batch-mode`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"MAVEN_REPO_URL":      "http://localhost:9100/test",
			"MAVEN_REPO_USERNAME": "admin",
			"MAVEN_REPO_PASSWORD": "secret",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
