package java

import (
	"github.com/EnvCLI/normalize-ci/pkg/common"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"strings"
)

// Action implementation
type RunActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n RunActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n RunActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n RunActionStruct) GetVersion() string {
	return n.version
}

// Check if this package can handle the current environment
func (n RunActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (n RunActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	buildSystem := DetectJavaBuildSystem(projectDir)
	if buildSystem == "gradle" {
		common.SetEnvironment(env, `GRADLE_OPTS`, `-XX:MaxMetaspaceSize=256m -XX:+HeapDumpOnOutOfMemoryError -Xmx512m`)

		command.RunCommand(`gradlew build --no-daemon --warning-mode=all --console=plain`, env)
	} else if buildSystem == "maven" {
		command.RunCommand(`mvn package -DskipTests=true`, env)
	} else {
		log.Fatal().Msg("can't detect build system")
	}

	files, filesErr := filesystem.FindFilesInDirectory(projectDir + `/build/libs`, `.jar`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", projectDir + `/build/libs`).Msg("failed to list files")
	}
	if len(files) == 1 {
		command.RunCommand(`java -jar ` + files[0] + ` ` + strings.Join(args, " "), env)
	} else {
		log.Warn().Int("count", len(files)).Msg("path build/libs should contain a single jar file! If you have a modular project please ensure that the final jar is moved into build/libs.")
	}
}

// RunAction
func RunAction() RunActionStruct {
	entity := RunActionStruct{
		stage: "run",
		name: "java-run",
		version: "0.1.0",
	}

	return entity
}
