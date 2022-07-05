package container

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"sort"
	"strings"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "container"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []*analyzerapi.ProjectModule {
	var result []*analyzerapi.ProjectModule

	// groovy
	files, filesErr := filesystem.FindFilesByExtension(ctx.ProjectDir, []string{})
	if filesErr == nil {
		// sort by length
		sort.Slice(files, func(i, j int) bool {
			return len(files[i]) < len(files[j])
		})

		// iterate
		for _, file := range files {
			filename := filepath.Base(file)

			if filename == "Dockerfile" {
				module := analyzerapi.ProjectModule{
					RootDirectory:     ctx.ProjectDir,
					Directory:         filepath.Dir(file),
					Name:              filepath.Base(filepath.Dir(file)),
					Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
					Discovery:         "file~" + file,
					BuildSystem:       analyzerapi.BuildSystemContainer,
					BuildSystemSyntax: nil,
					Language:          analyzerapi.GetSingleLanguageMap(analyzerapi.LanguageDockerfile, nil),
					Dependencies:      nil,
					Submodules:        nil,
					Files:             ctx.Files,
					FilesByExtension:  ctx.FilesByExtension,
				}
				analyzerapi.AddModuleToResult(&result, &module)
			} else if strings.HasSuffix(filename, ".sh") {
				content, contentErr := filesystem.GetFileContent(file)
				if contentErr == nil && strings.Contains(content, "buildah from") {
					module := analyzerapi.ProjectModule{
						RootDirectory:     ctx.ProjectDir,
						Directory:         filepath.Dir(file),
						Name:              filepath.Base(filepath.Dir(file)),
						Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
						Discovery:         "file~" + file,
						BuildSystem:       analyzerapi.BuildSystemContainer,
						BuildSystemSyntax: nil,
						Language:          analyzerapi.GetSingleLanguageMap(analyzerapi.LanguageBuildahScript, nil),
						Dependencies:      nil,
						Submodules:        nil,
						Files:             ctx.Files,
						FilesByExtension:  ctx.FilesByExtension,
					}
					analyzerapi.AddModuleToResult(&result, &module)
				} else if contentErr != nil {
					log.Warn().Str("file", file).Msg("failed to read file content")
				}
			}
		}
	}

	return result
}

func init() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, Analyzer{})
}
