package node

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzer_AnalyzeReact(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "react"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module
	log.Info().Interface("result", result).Msg("output")
}
