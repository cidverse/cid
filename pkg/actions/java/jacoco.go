package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

func processJacocoFile(ctx *api.ActionExecutionContext, file string) {
	// store jacoco test report if present
	jacocoReport := filepath.Join(ctx.CurrentModule.Directory, file)
	if filesystem.FileExists(jacocoReport) {
		content, contentErr := filesystem.GetFileContent(jacocoReport)
		if contentErr == nil {
			_ = filesystem.SaveFileText(filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug), "jacoco.xml"), content)
		}
	}
}
