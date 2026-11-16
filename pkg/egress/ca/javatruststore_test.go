package ca

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"
)

func TestWriteJavaTrustStore(t *testing.T) {
	// two roots with the same subject must both survive (unique aliases)
	first, second, inspection := certPEM(t), certPEM(t), certPEM(t)
	bundle := first + "garbage between blocks\n" + second + "-----BEGIN CERTIFICATE-----\nbroken\n-----END CERTIFICATE-----\n" + inspection
	path := filepath.Join(t.TempDir(), "truststore.p12")

	require.NoError(t, WriteJavaTrustStore(path, []byte(bundle)))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	certs, err := pkcs12.DecodeTrustStore(data, JavaTrustStorePassword)
	require.NoError(t, err)
	assert.Len(t, certs, 3, "broken entries are skipped")

	assert.Error(t, WriteJavaTrustStore(path, []byte("no certificates")))
}

func TestJavaToolOptions(t *testing.T) {
	options := "-Djavax.net.ssl.trustStore=/ts.p12 -Djavax.net.ssl.trustStoreType=PKCS12 -Djavax.net.ssl.trustStorePassword=changeit"
	assert.Equal(t, options, JavaToolOptions("", "/ts.p12"))
	assert.Equal(t, "-Xmx1g "+options, JavaToolOptions("  -Xmx1g ", "/ts.p12"))
}
