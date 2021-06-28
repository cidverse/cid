package gomod

import (
	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"golang.org/x/mod/modfile"
	"path/filepath"
	"sort"
)

type Analyzer struct{}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	// groovy
	files, filesErr := filesystem.FindFilesByExtension(ctx.ProjectDir, []string{".mod"})
	if filesErr == nil {
		// sort by length
		sort.Slice(files, func(i, j int) bool {
			return len(files[i]) < len(files[j])
		})

		// iterate
		for _, file := range files {
			filename := filepath.Base(file)

			// detect build system syntax
			if filename == "go.mod" {
				// parse go.mod
				contentBytes, contentReadErr := filesystem.GetFileBytes(file)
				if contentReadErr != nil {
					continue
				}
				goMod, goModParseError := modfile.ParseLax(file, contentBytes, nil)
				if goModParseError != nil {
					continue
				}

				// language
				language := make(map[analyzerapi.ProjectLanguage]*string)
				goversion := semver.MustParse(goMod.Go.Version).String()
				language[analyzerapi.LanguageGolang] = &goversion

				// deps
				var dependencies []analyzerapi.ProjectDependency
				for _, req := range goMod.Require {
					dep := analyzerapi.ProjectDependency{
						Type:    string(analyzerapi.BuildSystemGoMod),
						Id:      req.Mod.Path,
						Version: req.Mod.Version,
					}
					dependencies = append(dependencies, dep)
				}

				// module
				module := analyzerapi.ProjectModule{
					RootDirectory:     ctx.ProjectDir,
					Directory:         filepath.Dir(file),
					Name:              goMod.Module.Mod.Path,
					Slug:              slug.Make(goMod.Module.Mod.Path),
					Discovery:         "file~" + file,
					BuildSystem:       analyzerapi.BuildSystemGoMod,
					BuildSystemSyntax: nil,
					Language:          language,
					Dependencies:      dependencies,
					Submodules:        nil,
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