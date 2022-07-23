package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

func processJacocoFile(ctx *api.ActionExecutionContext, module *analyzerapi.ProjectModule, file string) {
	// store jacoco test report if present
	jacocoReport := filepath.Join(module.Directory, file)
	if filesystem.FileExists(jacocoReport) {
		content, contentErr := filesystem.GetFileContent(jacocoReport)
		if contentErr == nil {
			_ = filesystem.SaveFileText(filepath.Join(ctx.Paths.ArtifactModule(module.Slug, "test"), "jacoco.xml"), content)
		}
	}
}
