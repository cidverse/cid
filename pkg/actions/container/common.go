package container

import (
	"embed"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

//go:embed dockerfiles/*
var DockerfileFS embed.FS

// DetectAppType checks what kind of app the project is (via artifacts, should run after build actions)
func DetectAppType(ctx api.ActionExecutionContext) string {
	// java | jar
	files, filesErr := filesystem.FindFilesInDirectory(filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact), ".jar")
	if filesErr != nil {
		return ""
	}

	if len(files) > 0 {
		return "jar"
	}

	return ""
}
