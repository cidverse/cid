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
	binaryFile := filepath.Join(tmpDir, "linux_amd64")

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
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`}).Return([]*actionsdk.Artifact{
		{
			ArtifactID: "my-module|binary|linux_amd64",
			Module:     "my-module",
			Name:       "linux_amd64",
			Type:       "binary",
		},
	}, nil)
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{
		ID:         "my-module|binary|linux_amd64",
		TargetFile: binaryFile,
	}).Run(func(_ mock.Arguments) {
		require.NoError(t, os.WriteFile(binaryFile, []byte("binary content"), 0600))
	}).Return(&actionsdk.ArtifactDownloadResult{Path: binaryFile}, nil)
	var uploadReq *actionsdk.ArtifactUploadRequest

	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(req actionsdk.ArtifactUploadRequest) bool {
		if req.Type != "signature" {
			return false
		}
		uploadReq = &req
		return true
	})).Return("", "hash", nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())

	require.NotNil(t, uploadReq)
	assert.Equal(t, "signature", uploadReq.Type)
	assert.Equal(t, "pgp", uploadReq.Format)
	assert.Equal(t, "my-module", uploadReq.Module)
	assert.Equal(t, binaryFile+".asc", uploadReq.File)

	verifier, err := crypto.PGPWithProfile(profile.RFC9580()).Verify().
		VerificationKey(key).
		New()
	require.NoError(t, err)
	result, err := verifier.VerifyDetached([]byte("binary content"), uploadReq.ContentBytes, crypto.Armor)
	require.NoError(t, err)
	assert.NoError(t, result.SignatureError())
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
