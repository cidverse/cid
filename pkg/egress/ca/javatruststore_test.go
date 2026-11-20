package ca

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeJKS is a minimal jks truststore reader, mirroring what java checks when loading the store
func decodeJKS(t *testing.T, data []byte, password string) (aliases []string, certs []*x509.Certificate) {
	t.Helper()
	require.Greater(t, len(data), 20)
	body, sum := data[:len(data)-20], data[len(data)-20:]
	require.Equal(t, jksChecksum(password, body), sum, "integrity checksum")

	r := bytes.NewReader(body)
	read := func(v any) { require.NoError(t, binary.Read(r, binary.BigEndian, v)) }
	readUTF := func() string {
		var n uint16
		read(&n)
		b := make([]byte, n)
		read(b)
		return string(b)
	}

	var magic, version, count uint32
	read(&magic)
	read(&version)
	read(&count)
	require.Equal(t, uint32(jksMagic), magic)
	require.Equal(t, uint32(jksVersion), version)
	for range count {
		var tag uint32
		read(&tag)
		require.Equal(t, uint32(jksTrustedCertEntry), tag)
		aliases = append(aliases, readUTF())
		var ts int64
		read(&ts)
		require.Equal(t, "X.509", readUTF())
		var n uint32
		read(&n)
		der := make([]byte, n)
		read(der)
		cert, err := x509.ParseCertificate(der)
		require.NoError(t, err)
		certs = append(certs, cert)
	}
	require.Zero(t, r.Len(), "trailing data")
	return aliases, certs
}

func TestWriteJavaTrustStore(t *testing.T) {
	// two roots with the same subject must both survive (unique aliases), duplicates are dropped
	first, second, inspection := certPEM(t), certPEM(t), certPEM(t)
	bundle := first + "garbage between blocks\n" + second + "-----BEGIN CERTIFICATE-----\nbroken\n-----END CERTIFICATE-----\n" + inspection + first
	path := filepath.Join(t.TempDir(), "truststore.jks")

	require.NoError(t, WriteJavaTrustStore(path, []byte(bundle)))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	aliases, certs := decodeJKS(t, data, JavaTrustStorePassword)
	assert.Len(t, certs, 3, "broken and duplicate entries are skipped")
	assert.Len(t, aliases, 3)
	assert.NotEqual(t, aliases[0], aliases[1])

	assert.NotEqual(t, jksChecksum("wrong", data[:len(data)-20]), data[len(data)-20:])
	assert.Error(t, WriteJavaTrustStore(path, []byte("no certificates")))
}

func TestJKSAliasName(t *testing.T) {
	assert.Equal(t, "isrg-root-x1", jksAliasName("ISRG Root X1"))
	assert.Equal(t, "t-rkiye-ca", jksAliasName("  Türkiye CA "))
	assert.Equal(t, "acme-corp", jksAliasName("", "ACME Corp"))
	assert.Equal(t, "", jksAliasName(""))
}

func TestJavaToolOptions(t *testing.T) {
	options := "-Djavax.net.ssl.trustStore=/ts.jks -Djavax.net.ssl.trustStoreType=JKS -Djavax.net.ssl.trustStorePassword=changeit"
	assert.Equal(t, options, JavaToolOptions("", "/ts.jks"))
	assert.Equal(t, "-Xmx1g "+options, JavaToolOptions("  -Xmx1g ", "/ts.jks"))
}
