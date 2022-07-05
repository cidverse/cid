package analyzerapi

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func GetAnalyzerContext(projectDir string) AnalyzerContext {
	// files
	var files []string
	filesByExtension := make(map[string][]string)

	err := filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		// check for directory skip
		if d.IsDir() {
			if d.Name() == ".git" || path == filepath.Join(projectDir, "dist") || path == filepath.Join(projectDir, "tmp") || d.Name() == ".idea" {
				return filepath.SkipDir
			}
		} else {
			files = append(files, path)

			splitByExt := strings.SplitN(d.Name(), ".", 2)
			if len(splitByExt) == 2 {
				filesByExtension[splitByExt[1]] = append(filesByExtension[splitByExt[1]], path)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Str("path", projectDir).Msg("failed to retrieve directory contents")
	}

	// result
	return AnalyzerContext{
		ProjectDir:       projectDir,
		Files:            files,
		FilesByExtension: filesByExtension,
	}
}

func FindParentModule(modules []*ProjectModule, module *ProjectModule) *ProjectModule {
	for _, m := range modules {
		if strings.HasPrefix(module.Directory, m.Directory) {
			return m
		}
	}

	return nil
}

func GetSingleLanguageMap(language ProjectLanguage, version *string) map[ProjectLanguage]*string {
	languageMap := make(map[ProjectLanguage]*string)
	languageMap[language] = version
	return languageMap
}

func AddModuleToResult(result *[]*ProjectModule, module *ProjectModule) {
	parent := FindParentModule(*result, module)
	if parent != nil {
		module.Name = parent.Name + "-" + module.Name
		module.Slug = parent.Slug + "-" + module.Slug
		parent.Submodules = append(parent.Submodules, module)
	} else {
		*result = append(*result, module)
	}
}

func PrintStruct(t *testing.T, result interface{}) {
	jsonByteArray, jsonErr := json.MarshalIndent(result, "", "\t")
	assert.NoError(t, jsonErr)
	fmt.Println(string(jsonByteArray))
}
