package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseAssetNames(t *testing.T) {
	names := ReleaseAssetNames([]*actionsdk.Artifact{
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64"},
		{ArtifactID: "server|binary|linux_amd64", Module: "server", Name: "linux_amd64"},
		{ArtifactID: "cli|binary|darwin_arm64", Module: "cli", Name: "darwin_arm64"},
	})

	assert.Equal(t, map[string]string{
		"cli|binary|linux_amd64":    "cli_linux_amd64",
		"server|binary|linux_amd64": "server_linux_amd64",
		"cli|binary|darwin_arm64":   "darwin_arm64",
	}, names)
}

func TestReleaseAssetNamesSingleModule(t *testing.T) {
	names := ReleaseAssetNames([]*actionsdk.Artifact{
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64"},
		{ArtifactID: "cli|binary|windows_amd64.exe", Module: "cli", Name: "windows_amd64.exe"},
	})

	assert.Equal(t, map[string]string{
		"cli|binary|linux_amd64":       "linux_amd64",
		"cli|binary|windows_amd64.exe": "windows_amd64.exe",
	}, names)
}

func TestBinaryArtifacts(t *testing.T) {
	binaries := BinaryArtifacts([]*actionsdk.Artifact{
		{ArtifactID: "a", Type: "binary"},
		{ArtifactID: "b", Type: "signature"},
		{ArtifactID: "c", Type: "checksum"},
	})

	require.Len(t, binaries, 1)
	assert.Equal(t, "a", binaries[0].ArtifactID)
}

func TestSHA256FileAndChecksumLine(t *testing.T) {
	file := filepath.Join(t.TempDir(), "linux_amd64")
	require.NoError(t, os.WriteFile(file, []byte("hello\n"), 0o600))

	digest, err := SHA256File(file)
	require.NoError(t, err)
	assert.Equal(t, "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03", digest)
	assert.Equal(t, "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03  linux_amd64\n", ChecksumLine(digest, "linux_amd64"))
}
