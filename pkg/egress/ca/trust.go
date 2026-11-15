package ca

import (
	"bytes"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// DefaultTrustBundles are the system ca bundles the inspection ca is appended to (missing files are skipped)
var DefaultTrustBundles = []string{
	"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // fedora / rhel (podman/stable)
	"/etc/pki/tls/certs/ca-bundle.crt",                  // fedora / rhel (usually a symlink to the above)
	"/etc/ssl/certs/ca-certificates.crt",                // debian / ubuntu / alpine
	"/etc/ssl/cert.pem",                                 // alpine
}

// trustMarker starts the block appended to a trust bundle, everything from the marker to the end of the file is ours
const trustMarker = "# cid egress inspection ca"

// stripTrustBlock removes a ca block appended by a previous run
func stripTrustBlock(content []byte) ([]byte, bool) {
	idx := bytes.Index(content, []byte(trustMarker))
	if idx < 0 {
		return content, false
	}
	return content[:idx], true
}

// appendTrustBlock appends the ca certificate with the marker line to a bundle
func appendTrustBlock(content []byte, certPEM []byte) []byte {
	var buf bytes.Buffer
	buf.Write(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString(trustMarker + ", removed on shutdown\n")
	buf.Write(certPEM)
	return buf.Bytes()
}

// CleanupTrust removes ca blocks left behind by a previous run that was killed before it could restore the bundles
func CleanupTrust(bundles []string) error {
	var errs []error
	seen := map[string]bool{}
	for _, bundle := range bundles {
		real, err := filepath.EvalSymlinks(bundle)
		if err != nil || seen[real] {
			continue
		}
		seen[real] = true
		content, err := os.ReadFile(real)
		if err != nil {
			continue
		}
		if stripped, found := stripTrustBlock(content); found {
			log.Warn().Str("bundle", real).Msg("removing egress inspection ca left behind by a previous cid egress run")
			if err = os.WriteFile(real, stripped, 0o644); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// TrustInstall appends a ca certificate to system ca bundles and restores their original content on Restore
type TrustInstall struct {
	original map[string][]byte
	certPEM  []byte
	roots    []byte // content of the first system bundle without the ca
}

func InstallTrust(certPEM []byte, bundles []string) (*TrustInstall, error) {
	t := &TrustInstall{original: map[string][]byte{}, certPEM: certPEM}
	for _, bundle := range bundles {
		// bundles are often symlinks to each other, only modify each real file once
		real, err := filepath.EvalSymlinks(bundle)
		if err != nil {
			continue
		}
		if _, done := t.original[real]; done {
			continue
		}
		content, err := os.ReadFile(real)
		if err != nil {
			continue
		}
		content, _ = stripTrustBlock(content)
		if err = os.WriteFile(real, appendTrustBlock(content, certPEM), 0o644); err != nil {
			_ = t.Restore()
			return nil, fmt.Errorf("failed to add ca to %s: %w", real, err)
		}
		t.original[real] = content
		if t.roots == nil {
			t.roots = content
		}
		log.Info().Str("bundle", real).Msg("added egress inspection ca to trust bundle")
	}
	if len(t.original) == 0 {
		log.Warn().Strs("bundles", bundles).Msg("no system trust bundle found, clients must trust the inspection ca explicitly")
	}
	return t, nil
}

// WriteBundle writes a standalone bundle with the roots this environment trusts and the ca, for clients that don't use
// the job container's trust store (nested containers ship their own) and for `SSL_CERT_FILE` overrides.
//
// Roots: `$SSL_CERT_FILE` if set (a custom bundle, e.g. with local root cas for internal domains), otherwise the first
// system bundle, plus every certificate file in the `$SSL_CERT_DIR` directories. Duplicates are skipped.
func (t *TrustInstall) WriteBundle(path string, getenv func(string) string) error {
	roots := t.roots
	if file := getenv("SSL_CERT_FILE"); file != "" {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read SSL_CERT_FILE %s: %w", file, err)
		}
		roots, _ = stripTrustBlock(content)
	}

	var buf bytes.Buffer
	seen := map[string]bool{}
	addPEM := func(content []byte) {
		for {
			var block *pem.Block
			block, content = pem.Decode(content)
			if block == nil {
				return
			}
			if block.Type != "CERTIFICATE" || seen[string(block.Bytes)] {
				continue
			}
			seen[string(block.Bytes)] = true
			_ = pem.Encode(&buf, block)
		}
	}
	addPEM(roots)
	for _, dir := range filepath.SplitList(getenv("SSL_CERT_DIR")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Warn().Err(err).Str("dir", dir).Msg("failed to read SSL_CERT_DIR, its certificates are missing from the egress ca bundle")
			continue
		}
		for _, entry := range entries {
			content, err := os.ReadFile(filepath.Join(dir, entry.Name())) // follows the hash symlinks of c_rehash
			if err == nil {
				addPEM(content)
			}
		}
	}

	if err := os.WriteFile(path, appendTrustBlock(buf.Bytes(), t.certPEM), 0o644); err != nil {
		return fmt.Errorf("failed to write ca bundle %s: %w", path, err)
	}
	return nil
}

func (t *TrustInstall) Restore() error {
	var errs []error
	for path, content := range t.original {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			errs = append(errs, err)
		}
	}
	t.original = map[string][]byte{}
	return errors.Join(errs...)
}
