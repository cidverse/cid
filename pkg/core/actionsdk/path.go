package actionsdk

import (
	"path"
	"path/filepath"
)

var JoinSeparator = ""

func JoinPath(elem ...string) string {
	if JoinSeparator == "/" {
		return path.Join(elem...)
	}

	return filepath.Join(elem...)
}
