// Package sigstore signs artifacts keyless with Sigstore: an ephemeral key is certified by Fulcio for the CI job's
// OIDC identity and the signature is recorded in the Rekor transparency log. The result is a Sigstore bundle
// (`.sigstore.json`), verifiable with e.g. `cosign verify-blob --bundle <file>.sigstore.json ...`.
package sigstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/sign"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// Audience is the OIDC audience Fulcio expects.
	Audience = "sigstore"
	// GitHubActionsIssuer is the OIDC issuer of GitHub Actions ID tokens, used as certificate issuer on verification.
	GitHubActionsIssuer = "https://token.actions.githubusercontent.com"
	// GitLabComIssuer is the OIDC issuer of gitlab.com CI ID tokens, used as certificate issuer on verification.
	GitLabComIssuer = "https://gitlab.com"
	// IDTokenEnv is the environment variable a CI job provides a sigstore ID token in (GitLab job id_tokens).
	IDTokenEnv = "SIGSTORE_ID_TOKEN"
	// BundleExtension is the file extension of a Sigstore bundle published next to the signed file.
	BundleExtension = ".sigstore.json"
)

// IDToken returns the CI job's OIDC ID token for the sigstore audience: SIGSTORE_ID_TOKEN when the job provides it
// (GitLab CI id_tokens), otherwise requested from the GitHub Actions runtime.
func IDToken(ctx context.Context, client *http.Client, env map[string]string) (string, error) {
	if token := env[IDTokenEnv]; token != "" {
		return token, nil
	}
	if env["ACTIONS_ID_TOKEN_REQUEST_URL"] != "" || env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"] != "" {
		return GitHubActionsIDToken(ctx, client, env["ACTIONS_ID_TOKEN_REQUEST_URL"], env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"])
	}
	return "", errors.New("no OIDC ID token available: GitLab jobs need id_tokens " + IDTokenEnv + " (aud: " + Audience + "), GitHub jobs the id-token: write permission")
}

// GitHubActionsIDToken requests an OIDC ID token for the Sigstore audience from the GitHub Actions runtime.
// requestURL and requestToken are the runner-provided ACTIONS_ID_TOKEN_REQUEST_URL / ACTIONS_ID_TOKEN_REQUEST_TOKEN,
// only present in jobs with the `id-token: write` permission.
func GitHubActionsIDToken(ctx context.Context, client *http.Client, requestURL string, requestToken string) (string, error) {
	if requestURL == "" || requestToken == "" {
		return "", errors.New("ACTIONS_ID_TOKEN_REQUEST_URL / ACTIONS_ID_TOKEN_REQUEST_TOKEN not set, the job needs the id-token: write permission")
	}

	u, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("invalid ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	q := u.Query()
	q.Set("audience", Audience)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request github actions id token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("failed to read github actions id token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to request github actions id token: status %d", resp.StatusCode)
	}

	var tokenResponse struct {
		Value string `json:"value"`
	}
	if err = json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse github actions id token response: %w", err)
	}
	if tokenResponse.Value == "" {
		return "", errors.New("github actions id token response contains no token")
	}
	return tokenResponse.Value, nil
}

// Signer signs blobs keyless against the Sigstore public-good instance.
type Signer struct {
	idToken       string
	trustedRoot   *root.TrustedRoot
	signingConfig *root.SigningConfig
}

// NewSigner loads the public-good trusted root and signing config via TUF (cached in cacheDir) for signing with idToken.
func NewSigner(idToken string, cacheDir string) (*Signer, error) {
	tufOptions := tuf.DefaultOptions()
	tufOptions.CachePath = cacheDir
	tufClient, err := tuf.New(tufOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sigstore tuf client: %w", err)
	}

	trustedRoot, err := root.GetTrustedRoot(tufClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load sigstore trusted root: %w", err)
	}
	signingConfig, err := root.GetSigningConfig(tufClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load sigstore signing config: %w", err)
	}

	return &Signer{idToken: idToken, trustedRoot: trustedRoot, signingConfig: signingConfig}, nil
}

// SignBlob signs data with a fresh ephemeral key and returns the Sigstore bundle as JSON.
func (s *Signer) SignBlob(data []byte) ([]byte, error) {
	now := time.Now()
	opts := sign.BundleOptions{
		TrustedRoot: s.trustedRoot, // verifies the fulcio certificate and transparency log / timestamp responses
	}

	// certificate for the CI job's OIDC identity
	fulcioService, err := root.SelectService(s.signingConfig.FulcioCertificateAuthorityURLs(), sign.FulcioAPIVersions, now)
	if err != nil {
		return nil, fmt.Errorf("failed to select fulcio service: %w", err)
	}
	opts.CertificateProvider = sign.NewFulcio(&sign.FulcioOptions{
		BaseURL: fulcioService.URL,
		Timeout: 30 * time.Second,
		Retries: 1,
	})
	opts.CertificateProviderOptions = &sign.CertificateProviderOptions{IDToken: s.idToken}

	// transparency log
	rekorServices, err := root.SelectServices(s.signingConfig.RekorLogURLs(), s.signingConfig.RekorLogURLsConfig(), sign.RekorAPIVersions, now)
	if err != nil {
		return nil, fmt.Errorf("failed to select rekor service: %w", err)
	}
	for _, rekorService := range rekorServices {
		opts.TransparencyLogs = append(opts.TransparencyLogs, sign.NewRekor(&sign.RekorOptions{
			BaseURL: rekorService.URL,
			Timeout: 90 * time.Second,
			Retries: 1,
			Version: rekorService.MajorAPIVersion,
		}))
	}

	// signed timestamps (required alongside rekor v2, which no longer provides an integrated timestamp)
	tsaServices, err := root.SelectServices(s.signingConfig.TimestampAuthorityURLs(), s.signingConfig.TimestampAuthorityURLsConfig(), sign.TimestampAuthorityAPIVersions, now)
	if err != nil {
		return nil, fmt.Errorf("failed to select timestamp authority: %w", err)
	}
	for _, tsaService := range tsaServices {
		opts.TimestampAuthorities = append(opts.TimestampAuthorities, sign.NewTimestampAuthority(&sign.TimestampAuthorityOptions{
			URL:     tsaService.URL,
			Timeout: 30 * time.Second,
			Retries: 1,
		}))
	}

	keypair, err := sign.NewEphemeralKeypair(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ephemeral keypair: %w", err)
	}

	bundle, err := sign.Bundle(&sign.PlainData{Data: data}, keypair, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	bundleJSON, err := protojson.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sigstore bundle: %w", err)
	}
	return bundleJSON, nil
}
