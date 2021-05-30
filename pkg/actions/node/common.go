package node

import (
	"encoding/json"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

type PackageStruct struct {
	Name string
	Version string
	Dependencies map[string]string
	Scripts map[string]string
}

// DetectNodeProject checks if the target directory is a java project
func DetectNodeProject(projectDir string) bool {
	// package.json
	return filesystem.FileExists(projectDir+"/package.json")
}

// ParsePackageJSON will parse the package.json to evaluate its content
func ParsePackageJSON(file string) PackageStruct {
	var result PackageStruct

	// package.json
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		if err != nil {
			log.Debug().Err(err).Str("file", file).Msg("failed to open package.json")
			return result
		}

		fileBytes, fileErr := ioutil.ReadFile(file)
		if fileErr == nil {
			unmarshalErr := json.Unmarshal(fileBytes, &result)
			if unmarshalErr != nil {
				log.Fatal().Err(unmarshalErr).Str("file", file).Msg("failed to parse package.json")
			}
		} else {
			log.Debug().Err(fileErr).Str("file", file).Msg("failed to open package.json")
		}
	}

	return result
}
