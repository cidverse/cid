package api

import "embed"

func GetFileContentFromEmbedFS(fs embed.FS, file string) (string, error) {
	fileBytes, fileErr := fs.ReadFile(file)

	if fileErr != nil {
		return "", fileErr
	}

	return string(fileBytes), nil
}
