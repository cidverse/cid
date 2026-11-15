package egress

import (
	"os"
	"path/filepath"
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

	assert.Nil(t, caBundleEnv(bundle, withOverride), "no bundle written, the filter doesn't inspect tls")

	require.NoError(t, os.WriteFile(bundle, []byte("pem"), 0o644))
	assert.Equal(t, []string{"SSL_CERT_FILE=" + bundle}, caBundleEnv(bundle, withOverride))
	assert.Nil(t, caBundleEnv(bundle, noOverride), "system bundles are patched in place")
	assert.Nil(t, caBundleEnv("", withOverride))
}
