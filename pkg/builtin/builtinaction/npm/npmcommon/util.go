package npmcommon

import (
	"encoding/json"
)

type PackageJSON struct {
	Name         string
	Version      string
	Dependencies map[string]string
	Scripts      map[string]string
}

// ParsePackageJSON will parse the package.json to evaluate its content
func ParsePackageJSON(fileContent string) (result PackageJSON, err error) {
	err = json.Unmarshal([]byte(fileContent), &result)
	return
}
