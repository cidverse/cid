package util

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetUserConfigDirectory() string {
	if runtime.GOOS == "windows" {
		return getWindowsUserConfigDirectory()
	}

	return getUnixUserConfigDirectory()
}

func getWindowsUserConfigDirectory() string {
	cacheDir, _ := os.UserCacheDir()
	dir := filepath.Join(cacheDir, "cid")
	_ = os.MkdirAll(dir, os.ModePerm)

	return dir
}

func getUnixUserConfigDirectory() string {
	homeDir, _ := os.UserHomeDir()
	dir := filepath.Join(homeDir, ".config", "cid")
	_ = os.MkdirAll(dir, os.ModePerm)

	return dir
}
