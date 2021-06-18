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

func AnalyzeProject(projectDir string) []*analyzerapi.ProjectModule {
	if funk.Contains(analyticCache, projectDir) {
		return analyticCache[projectDir]
	}

	var result []*analyzerapi.ProjectModule

	for _, a := range analyzerapi.Analyzers {
		modules := a.Analyze(projectDir)
		for _, module := range modules {
			if !strings.Contains(module.Directory, "testdata") {
				result = append(result, module)
			}
		}
	}

	analyticCache[projectDir] = result
	return result
}
