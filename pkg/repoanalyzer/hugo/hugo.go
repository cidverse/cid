package hugo

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "hugo"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	// hugo
	for _, file := range append(ctx.FilesByExtension["toml"], ctx.FilesByExtension["yaml"]...) {
		filename := filepath.Base(file)
		if filename == "config.toml" || filename == "config.yaml" {
			hugoDir := filepath.Dir(file)
			if filesystem.DirectoryExists(filepath.Join(hugoDir, "content")) {
				// module
				module := analyzerapi.ProjectModule{
					RootDirectory:     ctx.ProjectDir,
					Directory:         filepath.Dir(file),
					Name:              filepath.Base(filepath.Dir(file)),
					Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
					Discovery:         []string{"file~" + file},
					BuildSystem:       analyzerapi.BuildSystemHugo,
					BuildSystemSyntax: analyzerapi.BuildSystemSyntaxDefault,
					Language:          nil,
					Dependencies:      nil,
					Submodules:        nil,
					Files:             ctx.Files,
					FilesByExtension:  ctx.FilesByExtension,
				}
				analyzerapi.AddModuleToResult(&result, &module)
			}
		}
	}

	return result
}
