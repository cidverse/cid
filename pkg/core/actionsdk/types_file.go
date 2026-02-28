package actionsdk

import (
	"path/filepath"
	"strings"
)

type FileV1Request struct {
	Directory  string   `json:"dir"`
	Extensions []string `json:"ext"`
}

type File struct {
	Path      string `json:"path"`
	Directory string `json:"dir"`
	Name      string `json:"name"`
	NameShort string `json:"name_short"`
	Extension string `json:"ext"`
}

func NewFile(path string) File {
	split := strings.SplitN(filepath.Base(path), ".", 2)
	fileName := split[0]
	fileExt := ""
	if len(split) > 1 && split[1] != "" {
		fileExt = "." + split[1]
	}

	return File{
		Path:      path,
		Directory: filepath.Dir(path),
		Name:      filepath.Base(path),
		NameShort: fileName,
		Extension: fileExt,
	}
}
