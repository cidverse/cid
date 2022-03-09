package hugo

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"path/filepath"
)

type Analyzer struct{}

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
					Discovery:         "file~" + file,
					BuildSystem:       analyzerapi.BuildSystemHugo,
					BuildSystemSyntax: nil,
					Language:          nil,
					Dependencies:      nil,
					Submodules:        nil,
					Files:             ctx.Files,
					FilesByExtension:  ctx.FilesByExtension,
				}
				result = append(result, &module)
			}
		}
	}

	return result
}

func init() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, Analyzer{})
}
