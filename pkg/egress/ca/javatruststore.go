package ca

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"software.sslmate.com/src/go-pkcs12"
)

// JavaTrustStorePassword protects the integrity of the java truststore only, it contains public certificates. Java
// requires a password for pkcs12 truststores on older releases, "changeit" is the JDK default.
const JavaTrustStorePassword = pkcs12.DefaultPassword

// JavaToolOptions returns the JVM options that make every JVM (incl. forked build daemons) use the truststore, appended
// to existing options. JVMs print "Picked up JAVA_TOOL_OPTIONS: ..." to stderr when the variable is set.
func JavaToolOptions(existing string, trustStore string) string {
	options := strings.Join([]string{
		"-Djavax.net.ssl.trustStore=" + trustStore,
		"-Djavax.net.ssl.trustStoreType=PKCS12",
		"-Djavax.net.ssl.trustStorePassword=" + JavaTrustStorePassword,
	}, " ")
	if existing = strings.TrimSpace(existing); existing != "" {
		return existing + " " + options
	}
	return options
}

// WriteJavaTrustStore writes the certificates of a pem bundle (see TrustInstall.WriteBundle) as pkcs12 java
// truststore. Java ignores the system pem bundles, it only trusts its own cacerts or the truststore passed via
// `-Djavax.net.ssl.trustStore` (see JavaToolOptions), which replaces cacerts - so the store contains all roots.
//
// Encoded with LegacyDES (3DES / SHA-1 mac), the only variant every JDK incl. 8 reads; the content is public anyway.
func WriteJavaTrustStore(path string, bundle []byte) error {
	var entries []pkcs12.TrustStoreEntry
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
		// unique aliases: roots can share a subject, java keeps only one entry per alias
		entries = append(entries, pkcs12.TrustStoreEntry{Cert: cert, FriendlyName: fmt.Sprintf("cid-%d-%s", len(entries), cert.Subject.CommonName)})
	}
	if len(entries) == 0 {
		return fmt.Errorf("no certificates for the java truststore")
	}

	data, err := pkcs12.LegacyDES.EncodeTrustStoreEntries(entries, JavaTrustStorePassword)
	if err != nil {
		return fmt.Errorf("failed to encode java truststore: %w", err)
	}
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write java truststore %s: %w", path, err)
	}
	return nil
}
