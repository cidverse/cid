package java

import (
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

type BuildActionStruct struct{}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "java-build",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// run build
	BuildJavaProject(ctx, state, ctx.CurrentModule)

	// colelct artifacts
	CollectGradleArtifacts(ctx, state, ctx.CurrentModule)

	// find artifacts
	files, _ := filesystem.FindFilesByExtension(ctx.ProjectDir, []string{".jar"})
	for _, file := range files {
		if strings.Contains(file, filepath.Join("build", "libs")) && IsJarExecutable(file) {
			moveErr := filesystem.MoveFile(files[0], filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug), filepath.Base(files[0])))
			log.Fatal().Err(moveErr).Msg("failed to move artifacts into artifact dir")
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}

func BuildJavaProject(ctx *api.ActionExecutionContext, state *api.ActionStateContext, module *analyzerapi.ProjectModule) {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]

	if module.BuildSystem == analyzerapi.BuildSystemGradle {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" assemble --no-daemon --warning-mode=all --console=plain`, ctx.Env, module.Directory)
	} else if module.BuildSystem == analyzerapi.BuildSystemMaven {
		MavenWrapperSetup(module.Directory)

		command.RunCommand(getMavenCommandPrefix(module.Directory)+" versions:set -DnewVersion="+releaseVersion+" --batch-mode", ctx.Env, module.Directory)
		command.RunCommand(getMavenCommandPrefix(module.Directory)+" package -DskipTests=true --batch-mode", ctx.Env, module.Directory)
	}
}
