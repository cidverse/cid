package util

import (
	"encoding/json"
	"os"
)

func StructToJsonFile(registry interface{}, file string) error {
	// marshal
	data, err := json.Marshal(&registry)
	if err != nil {
		return err
	}

	// write to filesystem
	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
