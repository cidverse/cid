package egress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCABundleEnv(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "bundle.crt")
	withOverride := func(k string) string {
		if k == "SSL_CERT_FILE" {
			return "/custom/roots.pem"
		}
		return ""
	}
	noOverride := func(string) string { return "" }

	assert.Nil(t, caBundleEnv(bundle, "", withOverride), "no bundle written, the filter doesn't inspect tls")

	require.NoError(t, os.WriteFile(bundle, []byte("pem"), 0o644))
	assert.Equal(t, []string{"SSL_CERT_FILE=" + bundle, "NODE_EXTRA_CA_CERTS=" + bundle, "REQUESTS_CA_BUNDLE=" + bundle}, caBundleEnv(bundle, "", withOverride))
	assert.Equal(t, []string{"SSL_CERT_FILE=" + bundle, "NODE_EXTRA_CA_CERTS=" + bundle, "REQUESTS_CA_BUNDLE=" + bundle}, caBundleEnv(bundle, "", noOverride), "clients with their own trust store ignore the patched system bundles")
	customNode := func(k string) string {
		if k == "NODE_EXTRA_CA_CERTS" {
			return "/custom/node.pem"
		}
		return ""
	}
	assert.Equal(t, []string{"SSL_CERT_FILE=" + bundle, "REQUESTS_CA_BUNDLE=" + bundle}, caBundleEnv(bundle, "", customNode), "custom values are kept")
	assert.Nil(t, caBundleEnv("", "", withOverride))
}

func TestCABundleEnvJavaTrustStore(t *testing.T) {
	dir := t.TempDir()
	bundle, trustStore := filepath.Join(dir, "bundle.crt"), filepath.Join(dir, "truststore.jks")
	require.NoError(t, os.WriteFile(bundle, []byte("pem"), 0o644))
	javaOptions := func(k string) string {
		if k == "JAVA_TOOL_OPTIONS" {
			return "-Xmx1g"
		}
		return ""
	}

	assert.NotContains(t, strings.Join(caBundleEnv(bundle, trustStore, javaOptions), "\n"), "JAVA_TOOL_OPTIONS", "no truststore written")

	require.NoError(t, os.WriteFile(trustStore, []byte("jks"), 0o644))
	env := caBundleEnv(bundle, trustStore, javaOptions)
	assert.Contains(t, env, "JAVA_TOOL_OPTIONS=-Xmx1g -Djavax.net.ssl.trustStore="+trustStore+" -Djavax.net.ssl.trustStoreType=JKS -Djavax.net.ssl.trustStorePassword=changeit", "existing options are kept")
}
