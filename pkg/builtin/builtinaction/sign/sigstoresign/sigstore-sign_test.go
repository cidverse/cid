package sigstoresign

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeSigner struct{}

func (fakeSigner) SignBlob(data []byte) ([]byte, error) {
	return []byte(`{"signed":"` + string(data) + `"}`), nil
}

func TestSigstoreSign(t *testing.T) {
	tmpDir := t.TempDir()
	data := common.TestProjectData()
	data.Config.TempDir = tmpDir
	data.Env["ACTIONS_ID_TOKEN_REQUEST_URL"] = "https://token.example/?api-version=2.0"
	data.Env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"] = "request-token"

	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(data, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: common.SignableArtifactQuery}).Return([]*actionsdk.Artifact{
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "server|binary|linux_amd64", Module: "server", Name: "linux_amd64", Type: "binary"},
		{ArtifactID: "cli|attestation|linux_amd64.provenance.json", Module: "cli", Name: "linux_amd64.provenance.json", Type: "attestation", Format: "provenance"},
	}, nil)
	for id, name := range map[string]string{"cli|binary|linux_amd64": "cli_linux_amd64", "server|binary|linux_amd64": "server_linux_amd64", "cli|attestation|linux_amd64.provenance.json": "cli_linux_amd64.provenance.json"} {
		file := filepath.Join(tmpDir, name)
		content := id
		sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{ID: id, TargetFile: file}).Run(func(_ mock.Arguments) {
			require.NoError(t, os.WriteFile(file, []byte(content), 0600))
		}).Return(&actionsdk.ArtifactDownloadResult{Path: file}, nil)
	}

	uploads := map[string]actionsdk.ArtifactUploadRequest{}
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(req actionsdk.ArtifactUploadRequest) bool {
		uploads[filepath.Base(req.File)] = req
		return true
	})).Return("", "hash", nil)

	var gotRequestURL, gotRequestToken, gotIDToken string
	action := Action{
		Sdk: sdk,
		idToken: func(_ context.Context, env map[string]string) (string, error) {
			gotRequestURL, gotRequestToken = env["ACTIONS_ID_TOKEN_REQUEST_URL"], env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"]
			return "id-token", nil
		},
		newSigner: func(idToken string, _ string) (blobSigner, error) {
			gotIDToken = idToken
			return fakeSigner{}, nil
		},
	}
	require.NoError(t, action.Execute())

	assert.Equal(t, "https://token.example/?api-version=2.0", gotRequestURL)
	assert.Equal(t, "request-token", gotRequestToken)
	assert.Equal(t, "id-token", gotIDToken)

	require.Len(t, uploads, 3)
	for name, want := range map[string]struct{ module, content string }{
		"cli_linux_amd64.sigstore.json":                 {"cli", `{"signed":"cli|binary|linux_amd64"}`},
		"server_linux_amd64.sigstore.json":              {"server", `{"signed":"server|binary|linux_amd64"}`},
		"cli_linux_amd64.provenance.json.sigstore.json": {"cli", `{"signed":"cli|attestation|linux_amd64.provenance.json"}`},
	} {
		req, ok := uploads[name]
		require.True(t, ok, "missing bundle %s", name)
		assert.Equal(t, "signature", req.Type)
		assert.Equal(t, "sigstore", req.Format)
		assert.Equal(t, want.module, req.Module)
		assert.Equal(t, want.content, string(req.ContentBytes))
	}
}

func TestSigstoreSignWithoutBinaries(t *testing.T) {
	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(common.TestProjectData(), nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: common.SignableArtifactQuery}).Return([]*actionsdk.Artifact{}, nil)

	action := Action{Sdk: sdk}
	require.NoError(t, action.Execute())
}

func TestMetadataRule(t *testing.T) {
	rule := catalog.WorkflowRule{
		Type:       catalog.WorkflowExpressionCEL,
		Expression: Action{}.Metadata().Rules[0].Expression,
	}
	eval := func(env map[string]string) bool {
		return rules.EvaluateRule(rule, rules.GetProjectRuleContext(env, nil))
	}
	github := "https://github.com/cidverse/cid.git"
	gitlab := "https://gitlab.com/cidverse/cid.git"

	// enabled by default on github.com / gitlab.com
	assert.True(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": github}))
	assert.True(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": gitlab}))
	assert.True(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": github, "CID_SIGSTORE_SIGN": "true"}))
	// public fulcio doesn't trust self-hosted instances
	assert.False(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": "https://gitlab.example.com/a/b.git"}))
	// opt-out
	assert.False(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": github, "CID_SIGSTORE_SIGN": "false"}))
	assert.False(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": gitlab, "CID_SIGSTORE_SIGN": "false"}))
	// a configured GPG key takes precedence (secrets are "***" in the render-time rule context)
	assert.False(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": github, "GPG_PRIVATE_KEY": "***"}))
	assert.False(t, eval(map[string]string{"NCI_REPOSITORY_REMOTE": gitlab, "GPG_PRIVATE_KEY": "***"}))
}

func TestMetadataRequestsSigstoreIDToken(t *testing.T) {
	assert.Equal(t, []actionsdk.ActionAccessIDToken{{Name: "SIGSTORE_ID_TOKEN", Audience: "sigstore"}}, Action{}.Metadata().Access.IDTokens)
}

func TestMetadataDoesNotRenderRunnerTokenEnv(t *testing.T) {
	for _, env := range (Action{}).Metadata().Access.Environment {
		if env.Name == "^ACTIONS_ID_TOKEN_REQUEST_URL$" || env.Name == "^ACTIONS_ID_TOKEN_REQUEST_TOKEN$" || env.Name == "^SIGSTORE_ID_TOKEN$" {
			assert.True(t, env.Pattern, "%s must be a pattern so the github workflow template doesn't override the runner value", env.Name)
		}
	}
}
