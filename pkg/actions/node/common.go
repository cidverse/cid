package node

import (
	"encoding/json"
	"os"

	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

type PackageStruct struct {
	Name         string
	Version      string
	Dependencies map[string]string
	Scripts      map[string]string
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

		fileContent, fileErr := filesystem.GetFileContent(file)
		if err == nil {
			unmarshalErr := json.Unmarshal([]byte(fileContent), &result)
			if unmarshalErr != nil {
				log.Fatal().Err(unmarshalErr).Str("file", file).Msg("failed to parse package.json")
			}
		} else {
			log.Debug().Err(fileErr).Str("file", file).Msg("failed to open package.json")
		}
	}

	return result
}
