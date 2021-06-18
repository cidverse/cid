package analyzerapi

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func FindParentModule(modules []*ProjectModule, module *ProjectModule) *ProjectModule {
	for _, m := range modules {
		if strings.HasPrefix(module.Directory, m.Directory) {
			return m
		}
	}

	return nil
}

func PrintStruct(t *testing.T, result interface{}) {
	jsonByteArray, jsonErr := json.MarshalIndent(result, "", "\t")
	assert.NoError(t, jsonErr)
	fmt.Println(string(jsonByteArray))
}
