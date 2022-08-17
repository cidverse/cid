package gradle

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/gosimple/slug"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "gradle"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	for _, file := range ctx.Files {
		filename := filepath.Base(file)

		// detect build system syntax
		var buildSystemSyntax analyzerapi.ProjectBuildSystemSyntax
		if filename == "build.gradle" {
			buildSystemSyntax = analyzerapi.GradleGroovyDSL
		} else if filename == "build.gradle.kts" {
			buildSystemSyntax = analyzerapi.GradleKotlinDSL
		} else {
			continue
		}

		// language
		language := make(map[analyzerapi.ProjectLanguage]*string)
		language[analyzerapi.LanguageJava] = nil

		// deps
		var dependencies []analyzerapi.ProjectDependency

		// module
		module := analyzerapi.ProjectModule{
			RootDirectory:     ctx.ProjectDir,
			Directory:         filepath.Dir(file),
			Name:              filepath.Base(filepath.Dir(file)),
			Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
			Discovery:         []string{"file~" + file},
			BuildSystem:       analyzerapi.BuildSystemGradle,
			BuildSystemSyntax: buildSystemSyntax,
			Language:          language,
			Dependencies:      dependencies,
			Submodules:        nil,
			Files:             ctx.Files,
			FilesByExtension:  ctx.FilesByExtension,
		}
		analyzerapi.AddModuleToResult(&result, &module)
	}

	return result
}
