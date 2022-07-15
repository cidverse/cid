package repoanalyzer

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/container"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/gomod"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/gradle"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/helm"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/hugo"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/node"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
	"strings"
	"time"
)

var analyzerCache = make(map[string][]*analyzerapi.ProjectModule)

// AnalyzeProject will analyze a project and return all modules in path
func AnalyzeProject(projectDir string, path string) []*analyzerapi.ProjectModule {
	if funk.Contains(analyzerCache, path) {
		return analyzerCache[path]
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
