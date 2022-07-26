package node

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
	"github.com/thoas/go-funk"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "node"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []analyzerapi.ProjectModule {
	var result []analyzerapi.ProjectModule

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
				typescriptVersion := packageData.Dependencies["typescript"]
				language[analyzerapi.LanguageTypescript] = &typescriptVersion
			}

			// deps
			var dependencies []analyzerapi.ProjectDependency
			for key, value := range packageData.Dependencies {
				dep := analyzerapi.ProjectDependency{
					Type:    string(analyzerapi.BuildSystemNpm),
					ID:      key,
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
				BuildSystemSyntax: analyzerapi.BuildSystemSyntaxDefault,
				Language:          language,
				Dependencies:      dependencies,
				Submodules:        nil,
				Files:             ctx.Files,
				FilesByExtension:  ctx.FilesByExtension,
			}

			analyzerapi.AddModuleToResult(&result, module)
		} else {
			continue
		}
	}

	return result
}
