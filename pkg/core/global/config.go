package global

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetUserConfigDirectory() string {
	if runtime.GOOS == "windows" {
		cacheDir, _ := os.UserCacheDir()
		dir := filepath.Join(cacheDir, "cid")
		_ = os.MkdirAll(dir, os.ModePerm)

		return dir
	} else {
		homeDir, _ := os.UserHomeDir()
		dir := filepath.Join(homeDir, ".cache", "cid")
		_ = os.MkdirAll(dir, os.ModePerm)

		return dir
	}
}
