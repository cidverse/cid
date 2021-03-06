package filesystem

import (
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// GetWorkingDirectory returns the current working directory
func GetWorkingDirectory() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Str("error", err.Error()).Msg("Couldn't detect working directory!")
	}

	return workingDir
}

// GetProjectDirectory will try to find the project directory based on repository folders (.git)
func GetProjectDirectory() (string, error) {
	currentDirectory := GetWorkingDirectory()
	var projectDirectory = ""
	log.Trace().Str("workingDirectory", currentDirectory).Msg("running GetProjectDirectory")

	directoryParts := strings.Split(currentDirectory, string(os.PathSeparator))

	for projectDirectory == "" {
		// git repository
		if _, err := os.Stat(filepath.Join(currentDirectory, "/.git")); err == nil {
			log.Trace().Str("projectDirectory", currentDirectory).Str("workingDirectory", GetWorkingDirectory()).Msg("found the project directory")
			return currentDirectory, nil
		}

		// cancel at root path
		if directoryParts[0]+"\\" == currentDirectory || currentDirectory == "/" {
			return "", errors.New("didn't find any repositories for the current working directory")
		}

		currentDirectory = filepath.Dir(currentDirectory)
		log.Trace().Str("currentDirectory", currentDirectory).Msg("proceed to search next directory")
	}

	return "", errors.New("didn't find any repositories for the current working directory")
}

func FindFilesInDirectory(directory string, extension string) ([]string, error) {
	var files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if len(extension) > 0 {
			if strings.HasSuffix(path, extension) {
				files = append(files, path)
			}
		} else {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
