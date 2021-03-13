package node

import (
	"github.com/rs/zerolog/log"
	"os"
)

// DetectNodeProject checks if the target directory is a java project
func DetectNodeProject(projectDir string) bool {
	// package.json
	if _, err := os.Stat(projectDir+"/package.json"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/package.json").Msg("found package.json")
		return true
	}

	return false
}
