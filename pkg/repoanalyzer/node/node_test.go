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

	analyzer := Analyzer{}
	result := analyzer.Analyze(filepath.Join(filepath.Dir(cwd), "testdata", "react"))

	// module

	// print result
	analyzerapi.PrintStruct(t, result)
}
