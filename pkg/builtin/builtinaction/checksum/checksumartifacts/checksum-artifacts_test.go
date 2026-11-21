package checksumartifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func sha(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func TestChecksumArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	data := common.TestProjectData()
	data.Config.TempDir = tmpDir

	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(data, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{
		{ArtifactID: "server|binary|linux_amd64", Module: "server", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "cli|binary|darwin_arm64", Module: "cli", Name: "darwin_arm64", Type: "binary"},
	}, nil)

	downloads := map[string]struct{ target, content string }{
		"server|binary|linux_amd64": {"server_linux_amd64", "server linux"},
		"cli|binary|linux_amd64":    {"cli_linux_amd64", "cli linux"},
		"cli|binary|darwin_arm64":   {"darwin_arm64", "cli darwin"},
	}
	for id, dl := range downloads {
		file := filepath.Join(tmpDir, dl.target)
		content := dl.content
		sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{ID: id, TargetFile: file}).Run(func(_ mock.Arguments) {
			require.NoError(t, os.WriteFile(file, []byte(content), 0600))
		}).Return(&actionsdk.ArtifactDownloadResult{Path: file}, nil)
	}

	uploads := map[string]actionsdk.ArtifactUploadRequest{}
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(req actionsdk.ArtifactUploadRequest) bool {
		uploads[filepath.Base(req.File)] = req
		return true
	})).Return("", "hash", nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())

	expected := map[string]struct{ module, content string }{
		"server_linux_amd64.sha256": {"server", sha("server linux") + "  server_linux_amd64\n"},
		"cli_linux_amd64.sha256":    {"cli", sha("cli linux") + "  cli_linux_amd64\n"},
		"darwin_arm64.sha256":       {"cli", sha("cli darwin") + "  darwin_arm64\n"},
		"SHA256SUMS": {"root", sha("cli linux") + "  cli_linux_amd64\n" +
			sha("cli darwin") + "  darwin_arm64\n" +
			sha("server linux") + "  server_linux_amd64\n"},
	}
	require.Len(t, uploads, len(expected))
	for name, want := range expected {
		req, ok := uploads[name]
		require.True(t, ok, "missing upload %s", name)
		assert.Equal(t, "checksum", req.Type)
		assert.Equal(t, "sha256", req.Format)
		assert.Equal(t, want.module, req.Module)
		assert.Equal(t, want.content, string(req.ContentBytes), name)
	}
}

func TestChecksumArtifactsWithoutBinaries(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{}, nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())
}
