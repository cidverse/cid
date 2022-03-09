package helm

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
	"path/filepath"
)

type Analyzer struct{}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	for _, file := range ctx.FilesByExtension["yaml"] {
		filename := filepath.Base(file)
		if filename == "Chart.yaml" {
			// module
			module := analyzerapi.ProjectModule{
				RootDirectory:     ctx.ProjectDir,
				Directory:         filepath.Dir(file),
				Name:              filepath.Base(filepath.Dir(file)),
				Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
				Discovery:         "file~" + file,
				BuildSystem:       analyzerapi.BuildSystemHelm,
				BuildSystemSyntax: nil,
				Language:          nil,
				Dependencies:      nil,
				Submodules:        nil,
				Files:             ctx.Files,
				FilesByExtension:  ctx.FilesByExtension,
			}

			parent := analyzerapi.FindParentModule(result, &module)
			if parent != nil {
				module.Name = parent.Name + "-" + module.Name
				module.Slug = parent.Slug + "-" + module.Slug
				parent.Submodules = append(parent.Submodules, &module)
			} else {
				result = append(result, &module)
			}
		}
	}

	return result
}

func init() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, Analyzer{})
}
