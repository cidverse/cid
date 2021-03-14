package container

import (
	"embed"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
)

//go:embed dockerfiles/*
var DockerfileFS embed.FS

// DetectAppType checks what kind of app the project is (via artifacts, should run after build actions)
func DetectAppType(projectDir string) string {
	// java | jar
	files, _ := filesystem.FindFilesInDirectory(projectDir+`/`+Config.Paths.Artifact, `.jar`)
	if len(files) > 0 {
		return "jar"
	}

	return ""
}

func GetFileContent(fs embed.FS, file string) (string, error) {
	fileBytes, fileErr := fs.ReadFile(file)

	if fileErr != nil {
		return "", fileErr
	}

	return string(fileBytes), nil
}
