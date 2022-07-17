package repoanalyzer

import (
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
	"github.com/thoas/go-funk"
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
	var result []*analyzerapi.ProjectModule
	for _, a := range analyzerapi.Analyzers {
		log.Debug().Str("name", a.GetName()).Msg("repo analyzer run")
		modules := a.Analyze(ctx)
		for _, module := range modules {
			if strings.HasPrefix(module.Directory, path) && !strings.Contains(module.Directory, "testdata") {
				result = append(result, module)
			}
		}
	}

	log.Info().Int("module_count", len(result)).Str("duration", time.Since(start).String()).Int("file_count", len(ctx.Files)).Msg("repo analyzer complete")

	analyzerCache[projectDir] = result
	return result
}

func initAnalyzers() {
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, container.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, gomod.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, gradle.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, helm.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, hugo.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, node.Analyzer{})
	analyzerapi.Analyzers = append(analyzerapi.Analyzers, python.Analyzer{})
}
