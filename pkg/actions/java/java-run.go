package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"strings"
)

type RunActionStruct struct{}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "run",
		Name:      "java-run",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action RunActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectJavaProject(ctx.ProjectDir)
}

// Execute runs the action
func (action RunActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	buildSystem := DetectJavaBuildSystem(ctx.ProjectDir)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		ctx.Env["GRADLE_OPTS"] = "-XX:MaxMetaspaceSize=256m -XX:+HeapDumpOnOutOfMemoryError -Xmx512m"

		command.RunCommand(GradleCommandPrefix+` build --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(ctx.ProjectDir)

		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" package -DskipTests=true --batch-mode", ctx.Env, ctx.ProjectDir)
	} else {
		log.Fatal().Msg("can't detect build system")
	}

	files, filesErr := filesystem.FindFilesInDirectory(ctx.ProjectDir+`/build/libs`, `.jar`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", ctx.ProjectDir+`/build/libs`).Msg("failed to list files")
	}
	if len(files) == 1 {
		_ = command.RunOptionalCommand(`java -jar `+files[0]+` `+strings.Join(ctx.Args, " "), ctx.Env, ctx.ProjectDir)
	} else {
		log.Warn().Int("count", len(files)).Msg("path build/libs should contain a single jar file! If you have a modular project please ensure that the final jar is moved into build/libs.")
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
