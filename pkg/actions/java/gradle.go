package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	cp "github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

// TODO: --add-opens=java.prefs/java.util.prefs=ALL-UNNAMED only on java9+
const GradleCommandPrefix = `java --add-opens=java.prefs/java.util.prefs=ALL-UNNAMED "-Dorg.gradle.appname=gradlew" "-classpath" "gradle/wrapper/gradle-wrapper.jar" "org.gradle.wrapper.GradleWrapperMain"`

func CollectGradleArtifacts(ctx *api.ActionExecutionContext, localState *state.ActionStateContext, module *analyzerapi.ProjectModule) {
	// collect
	classesSourceDir := filepath.Join(module.Directory, "build", "classes")
	classesTargetDir := ctx.Paths.ArtifactModule(module.Slug, "classes")

	if filesystem.DirectoryExists(classesSourceDir) {
		removeErr := os.RemoveAll(classesTargetDir)
		if removeErr != nil {
			log.Debug().Err(removeErr).Msg("failed to remove old classes")
		}
		copyErr := cp.Copy(classesSourceDir, classesTargetDir)
		if copyErr != nil {
			log.Err(copyErr).Msg("failed to copy generated classes")
		}
	}

	// recursion
	for _, submodule := range module.Submodules {
		CollectGradleArtifacts(ctx, localState, submodule)
	}
}
