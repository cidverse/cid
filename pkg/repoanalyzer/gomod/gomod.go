package gomod

import (
	"github.com/cidverse/cid/pkg/core/version"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"golang.org/x/mod/modfile"
	"path/filepath"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "gomod"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	for _, file := range ctx.FilesByExtension["mod"] {
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

			// references
			goVersion := version.Format(goMod.Go.Version)

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
				BuildSystemSyntax: analyzerapi.BuildSystemSyntaxDefault,
				Language:          analyzerapi.GetSingleLanguageMap(analyzerapi.LanguageGolang, &goVersion),
				Dependencies:      dependencies,
				Submodules:        nil,
				Files:             ctx.Files,
				FilesByExtension:  ctx.FilesByExtension,
			}

			result = append(result, &module)
		}
	}

	return result
}

func init() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, Analyzer{})
}
