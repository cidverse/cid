package container

import (
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
)

type Analyzer struct{}

func (a Analyzer) GetName() string {
	return "container"
}

func (a Analyzer) Analyze(ctx analyzerapi.AnalyzerContext) []analyzerapi.ProjectModule {
	var result []analyzerapi.ProjectModule

	// dockerfile
	for _, file := range ctx.Files {
		filename := filepath.Base(file)

		if filename == "Dockerfile" || filename == "Containerfile" || strings.HasSuffix(filename, ".Dockerfile") || strings.HasSuffix(filename, ".Containerfile") {
			module := analyzerapi.ProjectModule{
				RootDirectory:     ctx.ProjectDir,
				Directory:         filepath.Dir(file),
				Name:              filepath.Base(filepath.Dir(file)),
				Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
				Discovery:         "file~" + file,
				BuildSystem:       analyzerapi.BuildSystemContainer,
				BuildSystemSyntax: analyzerapi.ContainerFile,
				Language:          nil,
				Dependencies:      nil,
				Submodules:        nil,
				Files:             ctx.Files,
				FilesByExtension:  ctx.FilesByExtension,
			}
			analyzerapi.AddModuleToResult(&result, module)
		}
	}

	// buildah
	for _, file := range ctx.FilesByExtension["sh"] {
		filename := filepath.Base(file)

		if strings.HasSuffix(filename, ".sh") {
			content, contentErr := filesystem.GetFileContent(file)
			if contentErr != nil {
				log.Err(contentErr).Str("file", file).Msg("failed to read file")
			}
			if contentErr == nil && strings.Contains(content, "buildah from") {
				module := analyzerapi.ProjectModule{
					RootDirectory:     ctx.ProjectDir,
					Directory:         filepath.Dir(file),
					Name:              filepath.Base(filepath.Dir(file)),
					Slug:              slug.Make(filepath.Base(filepath.Dir(file))),
					Discovery:         "file~" + file,
					BuildSystem:       analyzerapi.BuildSystemContainer,
					BuildSystemSyntax: analyzerapi.ContainerBuildahScript,
					Language:          nil,
					Dependencies:      nil,
					Submodules:        nil,
					Files:             ctx.Files,
					FilesByExtension:  ctx.FilesByExtension,
				}
				analyzerapi.AddModuleToResult(&result, module)
			} else if contentErr != nil {
				log.Warn().Str("file", file).Msg("failed to read file content")
			}
		}
	}

	return result
}
