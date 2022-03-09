package gradle

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"path/filepath"
	"sort"
)

type Analyzer struct{}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	// groovy
	files, filesErr := filesystem.FindFilesByExtension(ctx.ProjectDir, []string{".gradle", ".gradle.kts"})
	if filesErr == nil {
		// sort by length
		sort.Slice(files, func(i, j int) bool {
			return len(files[i]) < len(files[j])
		})

		// iterate
		for _, file := range files {
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
				Discovery:         "file~" + file,
				BuildSystem:       analyzerapi.BuildSystemGradle,
				BuildSystemSyntax: &buildSystemSyntax,
				Language:          language,
				Dependencies:      dependencies,
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
