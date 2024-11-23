package util

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

func CIDConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "cid")
}

func CIDStateDir() string {
	return filepath.Join(xdg.StateHome, "cid")
}
