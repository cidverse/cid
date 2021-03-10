package python

import (
	"github.com/rs/zerolog/log"
	"os"
)

// DetectPythonProject checks if the target directory contains a supported python project
func DetectPythonProject(projectDir string) bool {
	if len(DetectPythonBuildSystem(projectDir)) > 0 {
		return true
	}

	return false
}

// DetectPythonBuildSystem returns the build system used in the project
func DetectPythonBuildSystem(projectDir string) string {
	// requirements.txt
	if _, err := os.Stat(projectDir+"/requirements.txt"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/requirements.txt").Msg("found requirements.txt project")
		return "requirements.txt"
	}

	// pipenv
	if _, err := os.Stat(projectDir+"/Pipfile"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/Pipfile").Msg("found pipenv project")
		return "pipenv"
	}

	// setup.py
	if _, err := os.Stat(projectDir+"/setup.py"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/setup.py").Msg("found setup.py project")
		return "setup.py"
	}

	return ""
}