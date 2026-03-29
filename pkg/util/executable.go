package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func GetExecutableHash() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
