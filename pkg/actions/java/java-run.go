package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"strings"
)

// Action implementation
type RunActionStruct struct {}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "run",
		Name: "java-run",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
	loadConfig(projectDirectory)

	buildSystem := DetectJavaBuildSystem(projectDirectory)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		env["GRADLE_OPTS"] = "-XX:MaxMetaspaceSize=256m -XX:+HeapDumpOnOutOfMemoryError -Xmx512m"

		command.RunCommand(GradleCommandPrefix+` build --no-daemon --warning-mode=all --console=plain`, env, projectDirectory)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(projectDirectory)

		command.RunCommand(getMavenCommandPrefix(projectDirectory)+" package -DskipTests=true --batch-mode", env, projectDirectory)
	} else {
		log.Fatal().Msg("can't detect build system")
	}

	files, filesErr := filesystem.FindFilesInDirectory(projectDirectory + `/build/libs`, `.jar`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", projectDirectory + `/build/libs`).Msg("failed to list files")
	}
	if len(files) == 1 {
		_ = command.RunOptionalCommand(`java -jar `+files[0]+` `+strings.Join(args, " "), env, projectDirectory)
	} else {
		log.Warn().Int("count", len(files)).Msg("path build/libs should contain a single jar file! If you have a modular project please ensure that the final jar is moved into build/libs.")
	}
}

// RunAction
func RunAction() RunActionStruct {
	return RunActionStruct{}
}
