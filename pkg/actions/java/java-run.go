package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"strings"
)

type RunActionStruct struct{}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "java-run",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action RunActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && (ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle || ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven)
}

// Execute runs the action
func (action RunActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle {
		ctx.Env["GRADLE_OPTS"] = "-XX:MaxMetaspaceSize=256m -XX:+HeapDumpOnOutOfMemoryError -Xmx512m"

		command.RunCommand(GradleCommandPrefix+` build --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
	} else if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven {
		MavenWrapperSetup(ctx.ProjectDir)

		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" package -DskipTests=true --batch-mode", ctx.Env, ctx.ProjectDir)
	} else {
		log.Fatal().Msg("can't detect build system")
	}

	files, filesErr := filesystem.FindFilesByExtension(ctx.ProjectDir+`/build/libs`, []string{".jar"})
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
