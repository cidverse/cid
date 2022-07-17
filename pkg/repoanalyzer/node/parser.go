package node

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/cidverse/cidverseutils/pkg/filesystem"
)

type PackageStruct struct {
	Name         string
	Version      string
	Dependencies map[string]string
	Scripts      map[string]string
}

// ParsePackageJSON will parse the package.json to evaluate its content
func ParsePackageJSON(file string) (PackageStruct, error) {
	var result PackageStruct

	// package.json
	if !filesystem.FileExists(file) {
		return PackageStruct{}, errors.New("failed to open package.json")
	}

	fileBytes, fileErr := os.ReadFile(file)
	if fileErr == nil {
		unmarshalErr := json.Unmarshal(fileBytes, &result)
		if unmarshalErr != nil {
			return PackageStruct{}, errors.New("failed to parse package.json")
		}
	} else {
		return PackageStruct{}, errors.New("failed to open package.json")
	}

	return result, nil
}
