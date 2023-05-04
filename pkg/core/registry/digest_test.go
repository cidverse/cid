package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetArtifactDigest(t *testing.T) {
	digest, err := GetArtifactDigest("quay.io/cidverse/base-ubi:9.1.0-17")
	assert.NoError(t, err)
	assert.Equal(t, "sha256:ef454485d07da2e28caaaf019b033d57a2ded2433b6483367b8335541f74a59c", digest)
}
