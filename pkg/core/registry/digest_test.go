package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArtifactDigest(t *testing.T) {
	digest, err := GetArtifactDigest("ghcr.io/cidverse/build-go:1.25.0")
	assert.NoError(t, err)
	assert.Equal(t, "sha256:1b991e97bc1aed083633a1060ce7341be4094413c83778b390fecb2db2a8f7c5", digest)
}
