package registry

import (
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"
)

// IsOCI determines whether a URL is to be treated as an OCI URL
func IsOCI(url string) bool {
	return strings.HasPrefix(url, fmt.Sprintf("%s://", OCIScheme))
}

// ParseReference will replace + with _ in the tag
func ParseReference(raw string) (registry.Reference, error) {
	parts := strings.Split(raw, ":")
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
		tag := parts[len(parts)-1]

		if tag != "" {
			newTag := strings.ReplaceAll(tag, "+", "_")
			raw = strings.ReplaceAll(raw, tag, newTag)
		}
	}

	return registry.ParseReference(raw)
}
