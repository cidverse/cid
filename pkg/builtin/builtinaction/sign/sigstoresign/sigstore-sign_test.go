package sigstoresign

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeSigner wraps the statement instead of signing it, so tests can inspect it.
type fakeSigner struct{}

func (fakeSigner) SignStatement(statement []byte) ([]byte, error) {
	return append([]byte(`{"signed":`), append(statement, '}')...), nil
}

func TestSigstoreSign(t *testing.T) {
	tmpDir := t.TempDir()
	data := common.TestProjectData()
	data.Config.TempDir = tmpDir
	data.Env["ACTIONS_ID_TOKEN_REQUEST_URL"] = "https://token.example/?api-version=2.0"
	data.Env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"] = "request-token"

	cliSHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	buildProvenance, err := provenance.GenerateStatement(provenance.Subject{Name: "linux_amd64", SHA256: cliSHA, Module: "cli"}, provenance.Input{Step: provenance.Step{Slug: "go-build-cli"}})
	require.NoError(t, err)

	sdk := common.TestSetup(t)
	sdk.On("ProjectExecutionContextV1").Return(data, nil)
	sdk.On("ArtifactListV1", actionsdk.ArtifactListRequest{Query: common.SignableArtifactQuery}).Return([]*actionsdk.Artifact{
		{ArtifactID: "cli|binary|linux_amd64", Module: "cli", Name: "linux_amd64", Type: "binary", StepSlug: "go-build-cli", SHA256: cliSHA},
		{ArtifactID: "server|binary|linux_amd64", Module: "server", Name: "linux_amd64", Type: "binary", StepSlug: "go-build-server"}, // no recorded digest, hashed after download
		{ArtifactID: "cli|sbom|cli.cdx.json", Module: "cli", Name: "cli.cdx.json", Type: "sbom", Format: "cyclonedx-json", StepSlug: "syft-sbom", SHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
		{ArtifactID: "cli|attestation|linux_amd64.provenance.json", Module: "cli", Name: "linux_amd64.provenance.json", Type: "attestation", Format: "provenance"},
	}, nil)
	serverFile := filepath.Join(tmpDir, "server_linux_amd64")
	sdk.On("ArtifactDownloadV1", actionsdk.ArtifactDownloadRequest{ID: "server|binary|linux_amd64", TargetFile: serverFile}).Run(func(_ mock.Arguments) {
		require.NoError(t, os.WriteFile(serverFile, []byte("server"), 0600))
	}).Return(&actionsdk.ArtifactDownloadResult{Path: serverFile}, nil)
	sdk.On("ArtifactDownloadByteArrayV1", actionsdk.ArtifactDownloadByteArrayRequest{ID: "cli|attestation|linux_amd64.provenance.json"}).Return(&actionsdk.ArtifactDownloadByteArrayResult{Bytes: buildProvenance}, nil)

	var upload actionsdk.ArtifactUploadRequest
	sdk.On("ArtifactUploadV1", mock.Anything).Run(func(args mock.Arguments) {
		upload = args.Get(0).(actionsdk.ArtifactUploadRequest)
	}).Return("", "hash", nil).Once()

	var gotRequestURL, gotRequestToken, gotIDToken string
	action := Action{
		Sdk: sdk,
		idToken: func(_ context.Context, env map[string]string) (string, error) {
			gotRequestURL, gotRequestToken = env["ACTIONS_ID_TOKEN_REQUEST_URL"], env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"]
			return "id-token", nil
		},
		newSigner: func(idToken string, _ string) (statementSigner, error) {
			gotIDToken = idToken
			return fakeSigner{}, nil
		},
	}
	require.NoError(t, action.Execute())

	assert.Equal(t, "https://token.example/?api-version=2.0", gotRequestURL)
	assert.Equal(t, "request-token", gotRequestToken)
	assert.Equal(t, "id-token", gotIDToken)

	// one bundle for the whole release
	assert.Equal(t, common.ReleaseAttestationFileName, filepath.Base(upload.File))
	assert.Equal(t, common.ReleaseAttestationModule, upload.Module)
	assert.Equal(t, "signature", upload.Type)
	assert.Equal(t, "sigstore", upload.Format)

	var signed struct {
		Signed struct {
			PredicateType string `json:"predicateType"`
			Subject       []struct {
				Name   string            `json:"name"`
				Digest map[string]string `json:"digest"`
			} `json:"subject"`
			Predicate struct {
				BuildDefinition struct {
					ExternalParameters map[string]interface{} `json:"externalParameters"`
				} `json:"buildDefinition"`
			} `json:"predicate"`
		} `json:"signed"`
	}
	require.NoError(t, json.Unmarshal(upload.ContentBytes, &signed))
	serverDigest := sha256.Sum256([]byte("server"))
	serverSHA := hex.EncodeToString(serverDigest[:])
	assert.Equal(t, provenance.PredicateType, signed.Signed.PredicateType)
	subjects := map[string]string{}
	for _, subject := range signed.Signed.Subject {
		subjects[subject.Name] = subject.Digest["sha256"]
	}
	assert.Equal(t, map[string]string{
		"cli_linux_amd64":    cliSHA,
		"server_linux_amd64": serverSHA,
		"cli.cdx.json":       "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}, subjects, "binaries and sboms under their release asset names, the provenance itself is no subject")
	assert.Equal(t, "git+https://github.com/cidverse/normalizeci.git@refs/tags/v1.2.0", signed.Signed.Predicate.BuildDefinition.ExternalParameters["source"])
}

func TestSigstoreSignWithoutArtifacts(t *testing.T) {
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
