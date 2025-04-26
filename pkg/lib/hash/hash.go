package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HashFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	hashInString := hex.EncodeToString(hashInBytes)

	return hashInString, nil
}
