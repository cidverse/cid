package node

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
	"github.com/thoas/go-funk"
	"path/filepath"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "node"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	// iterate
	for _, file := range ctx.FilesByExtension["json"] {
		filename := filepath.Base(file)

		// detect build system syntax
		if filename == "package.json" {
			packageData, packageDataErr := ParsePackageJSON(file)
			if packageDataErr != nil {
				continue
			}

			// language
			language := make(map[analyzerapi.ProjectLanguage]*string)
			language[analyzerapi.LanguageJavascript] = nil

			// - typescript?
			if funk.Contains(packageData.Dependencies, "typescript") {
				// semver.MustParse(packageData.Dependencies["typescript"])
				language[analyzerapi.LanguageTypescript] = nil
			}

			// deps
			var dependencies []analyzerapi.ProjectDependency
			for key, value := range packageData.Dependencies {
				dep := analyzerapi.ProjectDependency{
					Type:    string(analyzerapi.BuildSystemNpm),
					Id:      key,
					Version: value,
				}
				dependencies = append(dependencies, dep)
			}

			// module
			module := analyzerapi.ProjectModule{
				RootDirectory:     ctx.ProjectDir,
				Directory:         filepath.Dir(file),
				Name:              packageData.Name,
				Slug:              slug.Make(packageData.Name),
				Discovery:         "file~" + file,
				BuildSystem:       analyzerapi.BuildSystemNpm,
				BuildSystemSyntax: nil,
				Language:          language,
				Dependencies:      dependencies,
				Submodules:        nil,
				Files:             ctx.Files,
				FilesByExtension:  ctx.FilesByExtension,
			}

			parent := analyzerapi.FindParentModule(result, &module)
			if parent != nil {
				parent.Submodules = append(parent.Submodules, &module)
			} else {
				result = append(result, &module)
			}
		} else {
			continue
		}
	}

	return result
}

func init() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, Analyzer{})
}
