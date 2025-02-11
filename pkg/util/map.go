package util

import (
	"os"
)

// MergeMaps merges multiple maps into one
func MergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func ResolveEnvMap(env map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range env {
		result[k] = os.ExpandEnv(v)
	}
	return result
}
