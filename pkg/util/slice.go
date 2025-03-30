package util

import (
	"slices"
)

// CompactAndSort will compact and sort the given slice of strings
func CompactAndSort(data []string) []string {
	slices.Sort(data)
	return slices.Compact(data)
}
