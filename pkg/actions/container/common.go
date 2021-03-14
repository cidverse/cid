package container

import (
	"os"
)

// DetectAppType checks what kind of app the project is (via artifacts, should run after build actions)
func DetectAppType(projectDir string) string {
	// java | jar
	if _, err := os.Stat(projectDir+"/build/libs"); !os.IsNotExist(err) {
		return "jar"
	}

	return ""
}
