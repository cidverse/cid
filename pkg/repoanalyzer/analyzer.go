package repoanalyzer

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/gomod"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/gradle"
	_ "github.com/cidverse/cid/pkg/repoanalyzer/node"
	"github.com/thoas/go-funk"
	"strings"
)

var analyticCache = make(map[string][]*analyzerapi.ProjectModule)

// AnalyzeProject will analyze a project and return all modules in path
func AnalyzeProject(projectDir string, path string) []*analyzerapi.ProjectModule {
	if funk.Contains(analyticCache, projectDir) {
		return analyticCache[projectDir]
	}

	var result []*analyzerapi.ProjectModule
	for _, a := range analyzerapi.Analyzers {
		modules := a.Analyze(projectDir)
		for _, module := range modules {
			if strings.HasPrefix(module.Directory, path) && !strings.Contains(module.Directory, "testdata") {
				result = append(result, module)
			}
		}
	}

	analyticCache[projectDir] = result
	return result
}
