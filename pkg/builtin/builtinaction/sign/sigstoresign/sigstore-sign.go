package sigstoresign

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/lib/sigstore"
	"github.com/cidverse/cid/pkg/util"
)

const URI = "builtin://actions/sigstore-sign"

// blobSigner signs data and returns a Sigstore bundle (JSON).
type blobSigner interface {
	SignBlob(data []byte) ([]byte, error)
}

type Action struct {
	Sdk actionsdk.SDKClient

	// test hooks, nil uses the real GitHub Actions token endpoint / Sigstore public-good instance
	idToken   func(ctx context.Context, env map[string]string) (string, error)
	newSigner func(idToken string, cacheDir string) (blobSigner, error)
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "sigstore-sign",
		Description: "Signs binary artifacts keyless with Sigstore, using the CI job's OIDC identity (GitHub Actions / GitLab CI), and stores a Sigstore bundle (<asset>.sigstore.json) per binary and provenance statement.",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeProject,
		Rules: []actionsdk.ActionRule{
			{
				// enabled by default on github.com / gitlab.com (public Fulcio trusts their OIDC issuers, not self-hosted GitLab),
				// opt-out with CID_SIGSTORE_SIGN=false. A configured GPG key (sign-artifacts) takes precedence: checked at plan
				// time only, declaring GPG_PRIVATE_KEY here would make the GitHub workflow template pass the key into this job.
				// The workflow config limits it to projects with binary producing build systems.
				Type:       "cel",
				Expression: `(hasPrefix(ENV["NCI_REPOSITORY_REMOTE"], "https://github.com/") || hasPrefix(ENV["NCI_REPOSITORY_REMOTE"], "https://gitlab.com/")) && getMapValue(ENV, "CID_SIGSTORE_SIGN") != "false" && getMapValue(ENV, "GPG_PRIVATE_KEY") == ""`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{
				{
					Name:        "CID_SIGSTORE_SIGN",
					Description: "Set to false to disable keyless Sigstore signing (also skipped when GPG_PRIVATE_KEY is set). Signing records the repository and workflow in the public Rekor transparency log.",
				},
				// runner-provided, declared as patterns so the github workflow template doesn't override them
				{
					Name:        "^ACTIONS_ID_TOKEN_REQUEST_URL$",
					Description: "GitHub Actions OIDC token endpoint (runner-provided with id-token: write).",
					Pattern:     true,
				},
				{
					Name:        "^ACTIONS_ID_TOKEN_REQUEST_TOKEN$",
					Description: "GitHub Actions OIDC token request credential (runner-provided with id-token: write).",
					Pattern:     true,
					Secret:      true,
				},
				{
					Name:        "^" + sigstore.IDTokenEnv + "$",
					Description: "GitLab CI OIDC ID token for the sigstore audience (job id_tokens).",
					Pattern:     true,
					Secret:      true,
				},
			},
			IDTokens: []actionsdk.ActionAccessIDToken{
				{Name: sigstore.IDTokenEnv, Audience: sigstore.Audience},
			},
			Network: []actionsdk.ActionAccessNetwork{
				{Host: "*.actions.githubusercontent.com:443"}, // github oidc token endpoint (host varies per runner)
				{Host: "tuf-repo-cdn.sigstore.dev:443"},       // trusted root / signing config
				{Host: "fulcio.sigstore.dev:443"},             // signing certificate
				{Host: "rekor.sigstore.dev:443"},              // transparency log (v1)
				{Host: "*.rekor.sigstore.dev:443"},            // transparency log (v2 shards)
				{Host: "timestamp.sigstore.dev:443"},          // timestamp authority
			},
		},
		Input: actionsdk.ActionInput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type: "binary",
				},
			},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   "signature",
					Format: "sigstore",
				},
			},
		},
	}
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectExecutionContextV1()
	if err != nil {
		return err
	}

	signable, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: common.SignableArtifactQuery})
	if err != nil {
		return err
	}
	if len(common.BinaryArtifacts(signable)) == 0 {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "no binary artifacts found, skipping sigstore signing"})
		return nil
	}

	// identity token for the job
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	idTokenFn := a.idToken
	if idTokenFn == nil {
		idTokenFn = func(ctx context.Context, env map[string]string) (string, error) {
			return sigstore.IDToken(ctx, &http.Client{Timeout: 30 * time.Second}, env)
		}
	}
	idToken, err := idTokenFn(ctx, d.Env)
	if err != nil {
		return err
	}
	// GitLab's SIGSTORE_ID_TOKEN is redacted as a declared secret, a token requested from GitHub Actions is not
	util.ProtectSecret(idToken)

	newSignerFn := a.newSigner
	if newSignerFn == nil {
		newSignerFn = func(idToken string, cacheDir string) (blobSigner, error) {
			return sigstore.NewSigner(idToken, cacheDir)
		}
	}
	signer, err := newSignerFn(idToken, actionsdk.JoinPath(d.Config.TempDir, "sigstore-tuf"))
	if err != nil {
		return err
	}

	// sign every binary and its provenance under its release asset name
	assetNames := common.ReleaseFileNames(signable)
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signing artifacts with sigstore", Context: map[string]interface{}{"artifact_count": len(signable)}})
	for _, artifact := range signable {
		assetName := assetNames[artifact.ArtifactID]
		targetFile := actionsdk.JoinPath(d.Config.TempDir, assetName)
		_, dlErr := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			return fmt.Errorf("failed to retrieve artifact %s: %w", artifact.ArtifactID, dlErr)
		}

		content, readErr := os.ReadFile(targetFile)
		if readErr != nil {
			return fmt.Errorf("failed to read artifact %s: %w", artifact.ArtifactID, readErr)
		}

		bundle, signErr := signer.SignBlob(content)
		if signErr != nil {
			return fmt.Errorf("failed to sign artifact %s: %w", artifact.ArtifactID, signErr)
		}

		_, _, uploadErr := a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:         targetFile + sigstore.BundleExtension,
			ContentBytes: bundle,
			Module:       artifact.Module,
			Type:         "signature",
			Format:       "sigstore",
		})
		if uploadErr != nil {
			return fmt.Errorf("failed to store sigstore bundle for %s: %w", artifact.ArtifactID, uploadErr)
		}

		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signed artifact", Context: map[string]interface{}{"artifact": assetName, "bundle": assetName + sigstore.BundleExtension}})
	}

	return nil
}
