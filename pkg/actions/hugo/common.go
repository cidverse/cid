package hugo

import (
	"github.com/rs/zerolog/log"
	"os"
)

// DetectHugoProject checks if the target directory is a hugo project
func DetectHugoProject(projectDir string) bool {
	// config.toml
	if _, err := os.Stat(projectDir + "/config.toml"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/config.toml").Msg("found config.toml")
		return true
	}

	// config.yaml
	if _, err := os.Stat(projectDir + "/config.yaml"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/config.yaml").Msg("found config.yaml")
		return true
	}

	return false
}
