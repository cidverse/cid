package gradle

import (
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGradleAnalyzer_AnalyzeGroovy(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "gradle-groovy"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module
	assert.Len(t, result, 1)
	assert.Equal(t, "gradle-groovy", result[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemGradle, result[0].BuildSystem)
	assert.Equal(t, string(analyzerapi.GradleGroovyDSL), string(*result[0].BuildSystemSyntax))
	assert.Nil(t, result[0].Language[analyzerapi.LanguageJava])

	// submodule
	assert.Len(t, result[0].Submodules, 1)
	assert.Equal(t, "api", result[0].Submodules[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemGradle, result[0].Submodules[0].BuildSystem)
	assert.Equal(t, string(analyzerapi.GradleGroovyDSL), string(*result[0].Submodules[0].BuildSystemSyntax))
	assert.Nil(t, result[0].Submodules[0].Language[analyzerapi.LanguageJava])

	// print result
	analyzerapi.PrintStruct(t, result)
}

func TestGradleAnalyzer_AnalyzeKotlin(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	ctx := analyzerapi.GetAnalyzerContext(filepath.Join(filepath.Dir(cwd), "testdata", "gradle-kotlin"))
	analyzer := Analyzer{}
	result := analyzer.Analyze(ctx)

	// module
	assert.Len(t, result, 1)
	assert.Equal(t, "gradle-kotlin", result[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemGradle, result[0].BuildSystem)
	assert.Equal(t, string(analyzerapi.GradleKotlinDSL), string(*result[0].BuildSystemSyntax))
	assert.Nil(t, result[0].Language[analyzerapi.LanguageJava])

	// submodule
	assert.Len(t, result[0].Submodules, 1)
	assert.Equal(t, "api", result[0].Submodules[0].Name)
	assert.Equal(t, analyzerapi.BuildSystemGradle, result[0].Submodules[0].BuildSystem)
	assert.Equal(t, string(analyzerapi.GradleKotlinDSL), string(*result[0].Submodules[0].BuildSystemSyntax))
	assert.Nil(t, result[0].Submodules[0].Language[analyzerapi.LanguageJava])

	// print result
	analyzerapi.PrintStruct(t, result)
}
