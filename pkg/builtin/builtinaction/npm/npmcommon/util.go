package npmcommon

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/core/actionsdk"
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

// InstallCommand returns the dependency install command, using npm ci for reproducible installs if a lockfile is present
func InstallCommand(sdk actionsdk.SDKClient, moduleDir string) string {
	if sdk.FileExistsV1(actionsdk.JoinPath(moduleDir, "package-lock.json")) || sdk.FileExistsV1(actionsdk.JoinPath(moduleDir, "npm-shrinkwrap.json")) {
		return `npm ci`
	}

	return `npm install`
}

// IsDependencyPath checks if the path is inside a node_modules directory
func IsDependencyPath(path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/node_modules/")
}
