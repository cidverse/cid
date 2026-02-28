package builtin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	cp "github.com/otiai10/copy"
)

func (sdk ActionSDK) FileReadV1(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (sdk ActionSDK) FileWriteV1(file string, content []byte) error {
	err := os.WriteFile(file, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (sdk ActionSDK) FileRemoveV1(file string) error {
	err := os.Remove(file)
	if err != nil {
		return err
	}

	return nil
}

func (sdk ActionSDK) FileRenameV1(old string, new string) error {
	err := os.Rename(old, new)
	if err != nil {
		return err
	}

	return nil
}

func (sdk ActionSDK) FileCopyV1(old string, new string) error {
	err := cp.Copy(old, new)

	return err
}

func (sdk ActionSDK) FileListV1(req actionsdk.FileV1Request) (files []actionsdk.File, err error) {
	err = filepath.Walk(req.Directory, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			if len(req.Extensions) > 0 {
				for _, ext := range req.Extensions {
					if strings.HasSuffix(path, ext) {
						files = append(files, actionsdk.NewFile(path))
						break
					}
				}
			} else {
				files = append(files, actionsdk.NewFile(path))
			}
		}

		return nil
	})

	return
}

func (sdk ActionSDK) FileExistsV1(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}
