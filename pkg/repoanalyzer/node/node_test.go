package node

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer_AnalyzeReact(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "react"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module

	// print result
	analyzerapi.PrintStruct(t, result)
}