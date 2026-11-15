package ca

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const systemBundle = "-----BEGIN CERTIFICATE-----\nsystem\n-----END CERTIFICATE-----\n"

func TestTrustBlockAppendAndStrip(t *testing.T) {
	ca, err := New()
	require.NoError(t, err)

	installed := appendTrustBlock([]byte(systemBundle), ca.CertPEM())
	assert.True(t, strings.HasPrefix(string(installed), systemBundle))
	assert.True(t, strings.HasSuffix(string(installed), string(ca.CertPEM())))
	assert.Equal(t, 1, strings.Count(string(installed), trustMarker))

	stripped, found := stripTrustBlock(installed)
	assert.True(t, found)
	assert.Equal(t, systemBundle, string(stripped))

	_, found = stripTrustBlock([]byte(systemBundle))
	assert.False(t, found)
}

func TestTrustBlockMissingTrailingNewline(t *testing.T) {
	installed := appendTrustBlock([]byte("no-newline"), []byte("PEM\n"))
	assert.Equal(t, "no-newline\n"+trustMarker+", removed on shutdown\nPEM\n", string(installed))
}

func TestCAIssuesLeafForHost(t *testing.T) {
	ca, err := New()
	require.NoError(t, err)

	leaf, err := ca.LeafFor("API.GitHub.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"api.github.com"}, leaf.Leaf.DNSNames)
	assert.NoError(t, leaf.Leaf.CheckSignatureFrom(ca.cert))

	again, err := ca.LeafFor("api.github.com")
	require.NoError(t, err)
	assert.Same(t, leaf, again, "leaf certificates are cached per host")
}

func certPEM(t *testing.T) string {
	t.Helper()
	c, err := New()
	require.NoError(t, err)
	return string(c.CertPEM())
}

func TestWriteBundleUsesEnvironmentRoots(t *testing.T) {
	dir := t.TempDir()
	systemRoot, customRoot, localRoot, inspection := certPEM(t), certPEM(t), certPEM(t), certPEM(t)

	customFile := filepath.Join(dir, "custom.pem")
	require.NoError(t, os.WriteFile(customFile, []byte(customRoot), 0o644))
	certDir := filepath.Join(dir, "certs")
	require.NoError(t, os.Mkdir(certDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "local-root.crt"), []byte(localRoot), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "duplicate.pem"), []byte(customRoot), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "README"), []byte("not a certificate"), 0o644))

	trust := &TrustInstall{roots: []byte(systemRoot), certPEM: []byte(inspection)}
	out := filepath.Join(dir, "bundle.crt")

	// system roots only
	require.NoError(t, trust.WriteBundle(out, func(string) string { return "" }))
	content, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(content), systemRoot)
	assert.True(t, strings.HasSuffix(string(content), inspection))

	// SSL_CERT_FILE replaces the system roots, SSL_CERT_DIR adds to them
	env := map[string]string{"SSL_CERT_FILE": customFile, "SSL_CERT_DIR": certDir}
	require.NoError(t, trust.WriteBundle(out, func(k string) string { return env[k] }))
	content, err = os.ReadFile(out)
	require.NoError(t, err)
	assert.NotContains(t, string(content), systemRoot)
	assert.Equal(t, 1, strings.Count(string(content), customRoot), "duplicates are skipped")
	assert.Contains(t, string(content), localRoot)
	assert.NotContains(t, string(content), "not a certificate")
	assert.True(t, strings.HasSuffix(string(content), inspection))

	env["SSL_CERT_FILE"] = filepath.Join(dir, "missing.pem")
	assert.Error(t, trust.WriteBundle(out, func(k string) string { return env[k] }))
}
