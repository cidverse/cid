package gomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
)

func TestGoModAnalyzer_Analyze(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "gomod"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module
	assert.Len(t, result, 1)
	assert.Equal(t, "github.com/dummymodule", result[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemGoMod, result[0].BuildSystem)
	assert.Equal(t, analyzerapi.BuildSystemSyntaxDefault, result[0].BuildSystemSyntax)
	assert.NotNil(t, result[0].Language[analyzerapi.LanguageGolang])
	assert.Equal(t, "1.16.0", *result[0].Language[analyzerapi.LanguageGolang])
	assert.Equal(t, "gomod", result[0].Dependencies[0].Type)
	assert.Equal(t, "github.com/Masterminds/semver/v3", result[0].Dependencies[0].ID)
	assert.Equal(t, "v3.1.1", result[0].Dependencies[0].Version)

	// submodule
	assert.Len(t, result[0].Submodules, 0)

	// print result
	analyzerapi.PrintStruct(t, result)
}
