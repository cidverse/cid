package executable

import "strings"

// PinnedReference returns the digest-pinned image reference, falling back to the image as declared when no
// digest is known or the image already carries one.
func PinnedReference(image string, digest string) string {
	if digest == "" || strings.Contains(image, "@") {
		return image
	}

	return image + "@" + digest
}

// NamedReference returns the human readable tag reference of an image (e.g. "repo:tag"), without the pinned
// digest. Images that are referenced by digest only have no named reference and yield an empty string.
func NamedReference(image string) string {
	if strings.Contains(image, "@") {
		return ""
	}

	return image
}
