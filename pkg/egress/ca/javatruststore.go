package ca

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // the jks integrity checksum is defined as sha-1
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"
)

// JavaTrustStorePassword protects the integrity of the java truststore only, it contains public certificates.
// "changeit" is the JDK default.
const JavaTrustStorePassword = "changeit"

// jks format constants, see sun.security.provider.JavaKeyStore
const (
	jksMagic            = 0xFEEDFEED
	jksVersion          = 2
	jksTrustedCertEntry = 2
	jksIntegritySalt    = "Mighty Aphrodite" // hardcoded in the JDK's JavaKeyStore
)

// JavaToolOptions returns the JVM options that make every JVM (incl. forked build daemons) use the truststore, appended
// to existing options. JVMs print "Picked up JAVA_TOOL_OPTIONS: ..." to stderr when the variable is set.
func JavaToolOptions(existing string, trustStore string) string {
	options := strings.Join([]string{
		"-Djavax.net.ssl.trustStore=" + trustStore,
		"-Djavax.net.ssl.trustStoreType=JKS",
		"-Djavax.net.ssl.trustStorePassword=" + JavaTrustStorePassword,
	}, " ")
	if existing = strings.TrimSpace(existing); existing != "" {
		return existing + " " + options
	}
	return options
}

// WriteJavaTrustStore writes the certificates of a pem bundle (see TrustInstall.WriteBundle) as jks java truststore.
// Java ignores the system pem bundles, it only trusts its own cacerts or the truststore passed via
// `-Djavax.net.ssl.trustStore` (see JavaToolOptions), which replaces cacerts - so the store contains all roots.
//
// JKS is used because a truststore holds only public certificates: the format is a flat list of DER certificates plus
// a sha-1 integrity checksum, every JDK incl. 8 reads it, and it needs no pkcs12 / asn.1 library.
func WriteJavaTrustStore(path string, bundle []byte) error {
	data, err := EncodeJavaTrustStore(bundle)
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write java truststore %s: %w", path, err)
	}
	return nil
}

// EncodeJavaTrustStore encodes the certificates of a pem bundle as jks truststore, protected by JavaTrustStorePassword.
// Broken and duplicate certificates are skipped.
func EncodeJavaTrustStore(bundle []byte) ([]byte, error) {
	var certs []*x509.Certificate
	seen := map[[32]byte]bool{}
	for {
		var block *pem.Block
		block, bundle = pem.Decode(bundle)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue // a broken entry in the system bundle must not break java builds
		}
		if sum := sha256.Sum256(cert.Raw); !seen[sum] {
			seen[sum] = true
			certs = append(certs, cert)
		}
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates for the java truststore")
	}

	var buf bytes.Buffer
	w := func(v any) { _ = binary.Write(&buf, binary.BigEndian, v) } // writes to a bytes.Buffer don't fail
	writeUTF := func(s string) {
		w(uint16(len(s)))
		buf.WriteString(s)
	}

	w(uint32(jksMagic))
	w(uint32(jksVersion))
	w(uint32(len(certs)))
	for i, cert := range certs {
		w(uint32(jksTrustedCertEntry))
		// unique aliases: roots can share a subject, java keeps only one entry per alias
		alias := fmt.Sprintf("cid-%03d", i)
		if name := jksAliasName(append([]string{cert.Subject.CommonName}, cert.Subject.Organization...)...); name != "" {
			alias += "-" + name
		}
		writeUTF(alias)
		w(cert.NotBefore.UnixMilli())
		writeUTF("X.509")
		w(uint32(len(cert.Raw)))
		buf.Write(cert.Raw)
	}
	buf.Write(jksChecksum(JavaTrustStorePassword, buf.Bytes()))
	return buf.Bytes(), nil
}

// jksChecksum is SHA-1(password as UTF-16BE || "Mighty Aphrodite" || data), checked by java when loading the store
func jksChecksum(password string, data []byte) []byte {
	h := sha1.New() //nolint:gosec // defined by the jks format
	for _, c := range utf16.Encode([]rune(password)) {
		_, _ = h.Write([]byte{byte(c >> 8), byte(c)})
	}
	_, _ = h.Write([]byte(jksIntegritySalt))
	_, _ = h.Write(data)
	return h.Sum(nil)
}

// jksAliasName reduces the first non-empty name to lowercase ascii (java lowercases jks aliases, and ascii keeps the
// plain length-prefixed string identical to java's modified utf-8)
func jksAliasName(names ...string) string {
	name := ""
	for _, n := range names {
		if strings.TrimSpace(n) != "" {
			name = n
			break
		}
	}
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
		if b.Len() >= 64 {
			break
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
