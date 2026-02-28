package util

import "github.com/hashicorp/go-version"

// ByVersion implements sort.Interface
type ByVersion []*version.Version

func (a ByVersion) Len() int           { return len(a) }
func (a ByVersion) Less(i, j int) bool { return a[i].Compare(a[j]) > 0 }
func (a ByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
