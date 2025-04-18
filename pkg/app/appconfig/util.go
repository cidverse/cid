package appconfig

import (
	"fmt"
)

func addChange[T comparable](changes *[]ChangeEntry, key, scope, field string, oldVal, newVal T) {
	if oldVal != newVal {
		*changes = append(*changes, ChangeEntry{
			Workflow: key,
			Scope:    scope,
			Message:  fmt.Sprintf("[%s] changed from [`%v`] to [`%v`]", field, oldVal, newVal),
		})
	}
}

func addSliceChange(changes *[]ChangeEntry, key, scope, field string, oldVal, newVal []string) {
	if !slicesEqual(oldVal, newVal) {
		*changes = append(*changes, ChangeEntry{
			Workflow: key,
			Scope:    scope,
			Message:  fmt.Sprintf("[%s] changed from [`%v`] to [`%v`]", field, oldVal, newVal),
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
