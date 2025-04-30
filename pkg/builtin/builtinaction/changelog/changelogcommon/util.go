package changelogcommon

import (
	"embed"
	"errors"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cidverseutils/filesystem"
)

// GetFileContent returns the file content from either the directory or the embedded filesystem in that order
func GetFileContent(folder string, fs embed.FS, file string) (string, error) {
	if filesystem.FileExists(cidsdk.JoinPath(folder, file)) {
		content, err := filesystem.GetFileContent(cidsdk.JoinPath(folder, file))
		if err != nil {
			return "", err
		}

		return content, nil
	}

	// look in internal fs
	content, contentErr := GetFileContentFromEmbedFS(fs, "templates/"+file)
	if contentErr == nil {
		return content, nil
	}

	return "", errors.New("can't find template file " + file)
}

func GetFileContentFromEmbedFS(fs embed.FS, file string) (string, error) {
	fileBytes, fileErr := fs.ReadFile(file)

	if fileErr != nil {
		return "", fileErr
	}

	return string(fileBytes), nil
}

func AddLinks(input string) string {
	return input
}
