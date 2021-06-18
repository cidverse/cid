package gomod

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGoModAnalyzer_Analyze(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	analyzer := Analyzer{}
	result := analyzer.Analyze(filepath.Join(filepath.Dir(cwd), "testdata", "gomod"))

	// module
	assert.Len(t, result, 1)
	assert.Equal(t, "github.com/dummymodule", result[0].Name)
	assert.Equal(t, analyzerapi.GoMod, result[0].BuildSystem)
	assert.Nil(t, result[0].BuildSystemSyntax)
	assert.NotNil(t, result[0].Language[analyzerapi.Golang])
	assert.Equal(t, "1.16.0", result[0].Language[analyzerapi.Golang].String())
	assert.Equal(t, "gomod", result[0].Dependencies[0].Type)
	assert.Equal(t, "github.com/Masterminds/semver/v3", result[0].Dependencies[0].Id)
	assert.Equal(t, "v3.1.1", result[0].Dependencies[0].Version)

	// submodule
	assert.Len(t, result[0].Submodules, 0)

	// print result
	analyzerapi.PrintStruct(t, result)
}
