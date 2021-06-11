package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// Action implementation
type BuildActionStruct struct{}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "build",
		Name:      "java-build",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectJavaProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext) {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]

	// run build
	buildSystem := DetectJavaBuildSystem(ctx.ProjectDir)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" assemble --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(ctx.ProjectDir)

		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" versions:set -DnewVersion="+releaseVersion+"--batch-mode", ctx.Env, ctx.ProjectDir)
		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" package -DskipTests=true --batch-mode", ctx.Env, ctx.ProjectDir)
	}

	// find artifacts
	files, _ := filesystem.FindFilesInDirectory(ctx.ProjectDir, `.jar`)
	for _, file := range files {
		if strings.Contains(file, "build"+string(os.PathSeparator)+"libs") && IsJarExecutable(file) {
			moveErr := filesystem.MoveFile(files[0], ctx.ProjectDir+`/dist/`+filepath.Base(files[0]))
			log.Fatal().Err(moveErr).Msg("failed to move artifacts into artifact dir")
		}
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
