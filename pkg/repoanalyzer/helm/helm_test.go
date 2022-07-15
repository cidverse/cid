package helm

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer_AnalyzeHugo(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "helm"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module
	assert.Len(t, result, 1)
	assert.Equal(t, "mychart", result[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemHelm, result[0].BuildSystem)
	assert.Equal(t, analyzerapi.BuildSystemSyntaxDefault, result[0].BuildSystemSyntax)
	assert.Nil(t, result[0].Language)

	// print result
	analyzerapi.PrintStruct(t, result)
}
