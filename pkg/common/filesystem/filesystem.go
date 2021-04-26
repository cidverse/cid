package filesystem

import (
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
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

func CreateFileWithContent(file string, data string) error {
	err := ioutil.WriteFile(file, []byte(data), 0)

	if err != nil {
		return err
	}

	return nil
}

func RemoveFile(file string) error {
	err := os.Remove(file)
	if err != nil {
		return err
	}

	return nil
}

func MoveFile(oldLocation string, newLocation string) error {
	log.Info().Str("oldLocation", oldLocation).Str("newLocation", newLocation).Msg("moving file")
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		return err
	}

	return nil
}

// GetFileBytes will retrieve the content of a file as bytes
func GetFileBytes(file string) ([]byte, error) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		fileBytes, fileErr := ioutil.ReadFile(file)
		if fileErr == nil {
			return fileBytes, nil
		} else {
			return nil, err
		}
	}

	return nil, errors.New("file does not exist")
}

// GetFileContent will retrieve the content of a file as text
func GetFileContent(file string) (string, error) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		fileBytes, fileErr := ioutil.ReadFile(file)
		if fileErr == nil {
			return string(fileBytes), nil
		} else {
			return "", err
		}
	}

	return "", errors.New("file does not exist")
}

// SaveFileContent will save a file with the provided content
func SaveFileContent(file string, content string) error {
	data := []byte(content)

	err := ioutil.WriteFile(file, data, 0)

	return err
}

// FileExists checks if the file exists and returns a boolean
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// FileContainsString will check if a file contains the string
func FileContainsString(file string, str string) bool {
	content, contentErr := GetFileContent(file)
	if contentErr != nil {
		return false
	}

	if strings.Contains(content, str) {
		return true
	}

	return false
}

// DownloadFile will download a url and store it in local filepath.
func DownloadFile(url string, filepath string) error {
	log.Debug().Str("url", url).Str("path", filepath).Msg("Downloading file ...")

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
