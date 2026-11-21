package signartifacts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	key, err := crypto.PGPWithProfile(profile.RFC9580()).KeyGeneration().
		AddUserId("CID Test", "cid@example.com").
		New().
		GenerateKey()
	require.NoError(t, err)
	privateKey, err := key.Armor()
	require.NoError(t, err)

	data := common.TestProjectData()
	data.Config.TempDir = tmpDir
	data.Env["GPG_PRIVATE_KEY"] = privateKey

	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(data, nil)
	// two modules produce linux_amd64, cli produces a unique darwin_arm64
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "server|binary|linux_amd64", Module: "server", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "cli|binary|darwin_arm64", Module: "cli", Name: "darwin_arm64", Type: "binary"},
	}, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "checksum" && name == "SHA256SUMS"`}).Return([]*actionsdk.Artifact{
		{ArtifactID: "root|checksum|SHA256SUMS", Module: "root", Name: "SHA256SUMS", Type: "checksum", Format: "sha256"},
	}, nil)

	contents := map[string]string{
		"cli|binary|linux_amd64":    "cli linux",
		"server|binary|linux_amd64": "server linux",
		"cli|binary|darwin_arm64":   "cli darwin",
		"root|checksum|SHA256SUMS":  "checksums",
	}
	expectedTargets := map[string]string{
		"cli|binary|linux_amd64":    "cli_linux_amd64",
		"server|binary|linux_amd64": "server_linux_amd64",
		"cli|binary|darwin_arm64":   "darwin_arm64",
		"root|checksum|SHA256SUMS":  "SHA256SUMS",
	}
	for id, name := range expectedTargets {
		file := filepath.Join(tmpDir, name)
		content := contents[id]
		sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{ID: id, TargetFile: file}).Run(func(_ mock.Arguments) {
			require.NoError(t, os.WriteFile(file, []byte(content), 0600))
		}).Return(&actionsdk.ArtifactDownloadResult{Path: file}, nil)
	}

	uploads := map[string]actionsdk.ArtifactUploadRequest{}
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(req actionsdk.ArtifactUploadRequest) bool {
		if req.Type != "signature" {
			return false
		}
		uploads[filepath.Base(req.File)] = req
		return true
	})).Return("", "hash", nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())

	verifier, err := crypto.PGPWithProfile(profile.RFC9580()).Verify().
		VerificationKey(key).
		New()
	require.NoError(t, err)

	expected := map[string]struct{ module, content string }{
		"cli_linux_amd64.asc":    {"cli", "cli linux"},
		"server_linux_amd64.asc": {"server", "server linux"},
		"darwin_arm64.asc":       {"cli", "cli darwin"},
		"SHA256SUMS.asc":         {"root", "checksums"},
	}
	require.Len(t, uploads, len(expected))
	for name, want := range expected {
		req, ok := uploads[name]
		require.True(t, ok, "missing signature %s", name)
		assert.Equal(t, "pgp", req.Format)
		assert.Equal(t, want.module, req.Module)

		result, err := verifier.VerifyDetached([]byte(want.content), req.ContentBytes, crypto.Armor)
		require.NoError(t, err)
		assert.NoError(t, result.SignatureError())
	}
}

func TestMetadataRequiresSigningKey(t *testing.T) {
	metadata := Action{}.Metadata()
	rule := catalog.WorkflowRule{
		Type:       catalog.WorkflowExpressionCEL,
		Expression: metadata.Rules[0].Expression,
	}

	withoutKey := rules.GetProjectRuleContext(map[string]string{}, nil)
	assert.False(t, rules.EvaluateRule(rule, withoutKey))

	withKey := rules.GetProjectRuleContext(map[string]string{"GPG_PRIVATE_KEY": "***"}, nil)
	assert.True(t, rules.EvaluateRule(rule, withKey))
}

func TestSignArtifactsWithoutKey(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())
}
