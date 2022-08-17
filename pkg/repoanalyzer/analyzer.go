package repoanalyzer

import (
	"github.com/thoas/go-funk"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cid/pkg/repoanalyzer/container"
	"github.com/cidverse/cid/pkg/repoanalyzer/gomod"
	"github.com/cidverse/cid/pkg/repoanalyzer/gradle"
	"github.com/cidverse/cid/pkg/repoanalyzer/helm"
	"github.com/cidverse/cid/pkg/repoanalyzer/hugo"
	"github.com/cidverse/cid/pkg/repoanalyzer/node"
	"github.com/cidverse/cid/pkg/repoanalyzer/python"
	"github.com/rs/zerolog/log"
)

var analyzerCache = make(map[string][]*analyzerapi.ProjectModule)

// AnalyzeProject will analyze a project and return all modules in path
func AnalyzeProject(projectDir string, path string) []*analyzerapi.ProjectModule {
	if funk.Contains(analyzerCache, path) {
		return analyzerCache[path]
	}

	if len(analyzerapi.Analyzers) == 0 {
		initAnalyzers()
	}

	start := time.Now()
	log.Info().Str("path", path).Int("scanners", len(analyzerapi.Analyzers)).Msg("repo analyzer start")

	// prepare context
	ctx := analyzerapi.GetAnalyzerContext(projectDir)

	// run
	var allModules []*analyzerapi.ProjectModule
	var allModuleNames []string
	for _, a := range analyzerapi.Analyzers {
		log.Debug().Str("name", a.GetName()).Msg("repo analyzer run")
		modules := a.Analyze(ctx)
		for _, module := range modules {
			currentModule := module
			if strings.HasPrefix(currentModule.Directory, path) && !strings.Contains(currentModule.Directory, "testdata") {
				allModules = append(allModules, currentModule)
				allModuleNames = append(allModuleNames, currentModule.Slug)
			}
		}
	}

	log.Info().Int("module_count", len(allModules)).Strs("modules", allModuleNames).Str("duration", time.Since(start).String()).Int("file_count", len(ctx.Files)).Msg("repo analyzer complete")

	analyzerCache[projectDir] = allModules
	return allModules
}

func initAnalyzers() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers,
		container.Analyzer{},
		gomod.Analyzer{},
		gradle.Analyzer{},
		helm.Analyzer{},
		hugo.Analyzer{},
		node.Analyzer{},
		python.Analyzer{},
	)
}
