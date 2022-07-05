package container

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDockerfileSyntax(t *testing.T) {
	content := `# syntax=docker.io/docker/dockerfile:1.3
# platforms=linux/amd64,linux/arm64

FROM alpine:latest
`

	assert.Equal(t, "docker.io/docker/dockerfile:1.3", getDockerfileSyntax(content))
}

func TestGetDockerfileTargetPlatforms(t *testing.T) {
	content := `# syntax=docker.io/docker/dockerfile:1.3
# platforms=linux/amd64,linux/arm64

FROM alpine:latest
`

	platforms := getDockerfileTargetPlatforms(content)

	assert.Equal(t, 2, len(platforms))
	assert.Equal(t, "linux", platforms[0].OS)
	assert.Equal(t, "amd64", platforms[0].Arch)
	assert.Equal(t, "linux", platforms[1].OS)
	assert.Equal(t, "arm64", platforms[1].Arch)
}
