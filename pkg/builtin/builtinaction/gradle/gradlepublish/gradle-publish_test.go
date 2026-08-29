package gradlepublish

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradlePublish(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION": "false",
		"MAVEN_REPO_URL":       "http://localhost:9100/test",
		"MAVEN_REPO_USERNAME":  "admin",
		"MAVEN_REPO_PASSWORD":  "secret",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/gradlew").Return(true)
	sdk.On("FileExistsV1", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="1.0.0" publish --no-daemon --warning-mode=all --console=plain --stacktrace`,
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

func TestGradlePublishReleaseRepository(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION":            "false",
		"MAVEN_REPO_URL":                  "http://localhost:9100/test",
		"MAVEN_REPO_HEADER_VALUE":         "generic-token",
		"MAVEN_REPO_RELEASE_URL":          "http://localhost:9101/releases",
		"MAVEN_REPO_RELEASE_HEADER_NAME":  "Deploy-Token",
		"MAVEN_REPO_RELEASE_HEADER_VALUE": "release-token",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/gradlew").Return(true)
	sdk.On("FileExistsV1", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="1.0.0" publish --no-daemon --warning-mode=all --console=plain --stacktrace`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"MAVEN_REPO_URL":          "http://localhost:9101/releases",
			"MAVEN_REPO_USERNAME":     "",
			"MAVEN_REPO_PASSWORD":     "",
			"MAVEN_REPO_HEADER_NAME":  "Deploy-Token",
			"MAVEN_REPO_HEADER_VALUE": "release-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGradlePublishSnapshotRepository(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION":             "false",
		"MAVEN_VERSION":                    "2.0.0-SNAPSHOT",
		"MAVEN_REPO_URL":                   "http://localhost:9100/test",
		"MAVEN_REPO_HEADER_VALUE":          "generic-token",
		"MAVEN_REPO_SNAPSHOT_URL":          "http://localhost:9101/snapshots",
		"MAVEN_REPO_SNAPSHOT_HEADER_NAME":  "Deploy-Token",
		"MAVEN_REPO_SNAPSHOT_HEADER_VALUE": "snapshot-token",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/gradlew").Return(true)
	sdk.On("FileExistsV1", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="2.0.0-SNAPSHOT" publish --no-daemon --warning-mode=all --console=plain --stacktrace`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"MAVEN_REPO_URL":          "http://localhost:9101/snapshots",
			"MAVEN_REPO_USERNAME":     "",
			"MAVEN_REPO_PASSWORD":     "",
			"MAVEN_REPO_HEADER_NAME":  "Deploy-Token",
			"MAVEN_REPO_HEADER_VALUE": "snapshot-token",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}

func TestGradlePublishWithHeaderAuth(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ModuleExecutionContextV1").Return(gradlecommon.GradleTestData(map[string]string{
		"WRAPPER_VERIFICATION":    "false",
		"MAVEN_REPO_URL":          "https://gitlab.com/api/v4/groups/primelib/-/packages/maven",
		"MAVEN_REPO_HEADER_NAME":  "Deploy-Token",
		"MAVEN_REPO_HEADER_VALUE": "deploy-token-value",
	}, false), nil)
	sdk.On("FileExistsV1", "/my-project/gradlew").Return(true)
	sdk.On("FileExistsV1", "/my-project/gradle/wrapper/gradle-wrapper.jar").Return(true)
	sdk.On("ExecuteCommandV1", actionsdk.ExecuteCommandV1Request{
		Command: `java -Dorg.gradle.appname="gradlew" -classpath "/my-project/gradle/wrapper/gradle-wrapper.jar" org.gradle.wrapper.GradleWrapperMain -Pversion="1.0.0" publish --no-daemon --warning-mode=all --console=plain --stacktrace`,
		WorkDir: "/my-project",
		Env: map[string]string{
			"MAVEN_REPO_URL":          "https://gitlab.com/api/v4/groups/primelib/-/packages/maven",
			"MAVEN_REPO_USERNAME":     "",
			"MAVEN_REPO_PASSWORD":     "",
			"MAVEN_REPO_HEADER_NAME":  "Deploy-Token",
			"MAVEN_REPO_HEADER_VALUE": "deploy-token-value",
		},
	}).Return(&actionsdk.ExecuteCommandV1Response{Code: 0}, nil)

	action := Action{Sdk: sdk}
	err := action.Execute()
	assert.NoError(t, err)
}
