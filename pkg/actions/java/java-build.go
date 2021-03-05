package java

import (
	"github.com/EnvCLI/normalize-ci/pkg/common"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type BuildActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n BuildActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n BuildActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n BuildActionStruct) GetVersion() string {
	return n.version
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	buildSystem := DetectJavaBuildSystem(projectDir)
	if buildSystem == "gradle" {
		common.SetEnvironment(env, `GRADLE_OPTS`, `-XX:MaxMetaspaceSize=256m -XX:+HeapDumpOnOutOfMemoryError -Xmx512m`)
		command.RunCommand(`gradlew clean assemble --no-daemon --warning-mode=all --console=plain`, env)
	} else if buildSystem == "maven" {
		command.RunCommand(`mvn versions:set -DnewVersion=$NCI_COMMIT_REF_RELEASE`, env)
		command.RunCommand(`mvn clean package -DskipTests=true`, env)
	}
}

// BuildAction
func BuildAction() BuildActionStruct {
	entity := BuildActionStruct{
		stage: "build",
		name: "java-build",
		version: "0.1.0",
	}

	return entity
}
