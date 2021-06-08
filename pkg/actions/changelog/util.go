package changelog

import (
	"embed"
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

// GetFileContent returns the file content from either the directory or the embedded filesystem in that order
func GetFileContent(folder string, fs embed.FS, file string) (string, error) {
	if filesystem.FileExists(filepath.Join(folder, file)) {
		content, contentErr := filesystem.GetFileContent(filepath.Join(folder, file))
		if contentErr != nil {
			panic(contentErr)
		}

		return content, nil
	}

	// look in internal fs
	content, contentErr := api.GetFileContentFromEmbedFS(fs, "templates/" + file)
	if contentErr == nil {
		return content, nil
	}

	return "", errors.New("can't find template file " + file)
}

func AddLinks(input string) string {

	return input
}
