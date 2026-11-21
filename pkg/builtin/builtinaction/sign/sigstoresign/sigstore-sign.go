package sigstoresign

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cid/pkg/lib/sigstore"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/normalizeci/pkg/envstruct"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
)

const URI = "builtin://actions/sigstore-sign"

// statementSigner signs an in-toto statement and returns a Sigstore bundle (JSON).
type statementSigner interface {
	SignStatement(statement []byte) ([]byte, error)
}

type Action struct {
	Sdk actionsdk.SDKClient

	// test hooks, nil uses the real GitHub Actions token endpoint / Sigstore public-good instance
	idToken   func(ctx context.Context, env map[string]string) (string, error)
	newSigner func(idToken string, cacheDir string) (statementSigner, error)
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "sigstore-sign",
		Description: "Generates one SLSA provenance statement covering every release binary and SBOM, signs it keyless with Sigstore using the CI job's OIDC identity (GitHub Actions / GitLab CI) and stores the bundle as " + common.ReleaseAttestationFileName + ".",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeProject,
		Rules: []actionsdk.ActionRule{
			{
				// enabled by default on github.com / gitlab.com (public Fulcio trusts their OIDC issuers, not self-hosted GitLab),
				// opt-out with CID_SIGSTORE_SIGN=false. A configured GPG key (sign-artifacts) takes precedence: checked at plan
				// time only, declaring GPG_PRIVATE_KEY here would make the GitHub workflow template pass the key into this job.
				// The workflow config limits it to release workflows of projects with binaries or SBOMs.
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
				// runner-provided, identify the builder in the provenance (provenance.builderID)
				{
					Name:        "^GITHUB_SERVER_URL$",
					Description: "GitHub server URL (runner-provided).",
					Pattern:     true,
				},
				{
					Name:        "^GITHUB_WORKFLOW_REF$",
					Description: "GitHub workflow reference (runner-provided).",
					Pattern:     true,
				},
				{
					Name:        "^CI_RUNNER_ID$",
					Description: "GitLab runner id (runner-provided).",
					Pattern:     true,
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
				{
					Type:   common.SBOMArtifactType,
					Format: common.SBOMFormatCycloneDX,
				},
				{
					Type:   common.SBOMArtifactType,
					Format: common.SBOMFormatSPDX,
				},
				// the per-binary provenance is stored by the build step next to its binary and arrives with the binary input
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

	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: common.SignableArtifactQuery})
	if err != nil {
		return err
	}
	var subjects, buildProvenance []*actionsdk.Artifact
	for _, artifact := range artifacts {
		if common.IsProvenance(artifact) {
			buildProvenance = append(buildProvenance, artifact)
		} else {
			subjects = append(subjects, artifact)
		}
	}
	if len(subjects) == 0 {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "no binary or sbom artifacts found, skipping sigstore signing"})
		return nil
	}

	// release statement: every binary and sbom under its release asset name, with the build steps' provenance merged in
	assetNames := common.ReleaseFileNames(subjects)
	releaseSubjects := make([]provenance.ReleaseSubject, 0, len(subjects))
	for _, artifact := range subjects {
		digest, digestErr := a.artifactSHA256(d, artifact, assetNames[artifact.ArtifactID])
		if digestErr != nil {
			return digestErr
		}
		releaseSubjects = append(releaseSubjects, provenance.ReleaseSubject{
			Name:   assetNames[artifact.ArtifactID],
			SHA256: digest,
			Module: artifact.Module,
			Step:   artifact.StepSlug,
			Type:   artifact.Type,
			Format: artifact.Format,
		})
	}
	buildStatements := make([][]byte, 0, len(buildProvenance))
	for _, artifact := range buildProvenance {
		content, dlErr := a.Sdk.ArtifactDownloadByteArrayV1(actionsdk.ArtifactDownloadByteArrayRequest{ID: artifact.ArtifactID})
		if dlErr != nil {
			return fmt.Errorf("failed to retrieve provenance %s: %w", artifact.ArtifactID, dlErr)
		}
		buildStatements = append(buildStatements, content.Bytes)
	}

	var spec nci.Spec
	if specErr := envstruct.EnvMapToStruct(&spec, d.Env); specErr != nil {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "failed to read nci spec, provenance may be incomplete", Context: map[string]interface{}{"error": specErr.Error()}})
	}
	statement, err := provenance.GenerateReleaseStatement(releaseSubjects, provenance.Input{NCI: spec, Env: d.Env}, buildStatements)
	if err != nil {
		return fmt.Errorf("failed to generate release provenance: %w", err)
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
		newSignerFn = func(idToken string, cacheDir string) (statementSigner, error) {
			return sigstore.NewSigner(idToken, cacheDir)
		}
	}
	signer, err := newSignerFn(idToken, actionsdk.JoinPath(d.Config.TempDir, "sigstore-tuf"))
	if err != nil {
		return err
	}

	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signing release provenance with sigstore", Context: map[string]interface{}{"subject_count": len(releaseSubjects), "build_provenance_count": len(buildStatements)}})
	bundle, err := signer.SignStatement(statement)
	if err != nil {
		return fmt.Errorf("failed to sign release provenance: %w", err)
	}

	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:         actionsdk.JoinPath(d.Config.TempDir, common.ReleaseAttestationFileName),
		ContentBytes: bundle,
		Module:       common.ReleaseAttestationModule,
		Type:         "signature",
		Format:       "sigstore",
	})
	if err != nil {
		return fmt.Errorf("failed to store sigstore bundle: %w", err)
	}

	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signed release provenance", Context: map[string]interface{}{"bundle": common.ReleaseAttestationFileName}})
	return nil
}

// artifactSHA256 returns the recorded digest of the artifact, or downloads and hashes it when none was recorded.
func (a Action) artifactSHA256(d *actionsdk.ProjectExecutionContextV1Response, artifact *actionsdk.Artifact, assetName string) (string, error) {
	if artifact.SHA256 != "" {
		return artifact.SHA256, nil
	}

	targetFile := actionsdk.JoinPath(d.Config.TempDir, assetName)
	if _, err := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{ID: artifact.ArtifactID, TargetFile: targetFile}); err != nil {
		return "", fmt.Errorf("failed to retrieve artifact %s: %w", artifact.ArtifactID, err)
	}
	digest, err := common.SHA256File(targetFile)
	if err != nil {
		return "", fmt.Errorf("failed to hash artifact %s: %w", artifact.ArtifactID, err)
	}
	return digest, nil
}
