package common

import cidsdk "github.com/cidverse/cid-sdk-go"

func MergeActionAccessNetwork(groups ...[]cidsdk.ActionAccessNetwork) []cidsdk.ActionAccessNetwork {
	var merged []cidsdk.ActionAccessNetwork
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return merged
}
