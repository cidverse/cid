package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/rs/zerolog/log"
)

var certsDir = filepath.Join(xdg.ConfigHome, "cid", "certs")

var (
	ErrFailedToReadBundleFile = errors.New("failed to read bundle file")
	ErrNoCertBundleFound      = errors.New("no ca bundle found")
	ErrFailedToCreateCertDir  = errors.New("failed to create cert dir")
	ErrFailedToWriteCertFile  = errors.New("failed to write cert file")
	ErrUnknownCertFileType    = errors.New("unknown cert file type")
)

// GetCertFileByType returns the cert file by type (ca-bundle, java-keystore)
func GetCertFileByType(certFileType string) (string, error) {
	var files []string

	// Export the machine CA certificates to a file
	err := ExportMachineCACertsToFile(filepath.Join(certsDir, "ca-bundle.crt"))
	if err != nil {
		return "", err
	}

	if certFileType == "ca-bundle" {
		files = append(files, filepath.Join(certsDir, "ca-bundle.crt"))
	} else if certFileType == "java-keystore" {
		files = append(files, filepath.Join(certsDir, "keystore.jks"))
	} else {
		return "", errors.Join(ErrUnknownCertFileType, fmt.Errorf("certFileType: %s", certFileType))
	}

	for _, file := range files {
		if _, err = os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("cert file not found")
}

// CaBundles paths on various systems, see https://go.dev/src/crypto/x509/root_linux.go
var CaBundles = [][]string{
	{"/etc/ssl/certs/ca-certificates.crt"},                                  // Debian/Ubuntu/Gentoo etc.
	{"/etc/pki/tls/certs/ca-bundle.crt", "/etc/pki/tls/certs/ca-extra.crt"}, // Fedora/RHEL 6
	{"/etc/ssl/ca-bundle.pem"},                                              // OpenSUSE
	{"/etc/pki/tls/cacert.pem"},                                             // OpenELEC
	{"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"},                   // CentOS/RHEL 7
	{"/etc/ssl/cert.pem"},                                                   // Alpine Linux
}

// ExportMachineCACertsToFile exports the machine CA certificates to a file
func ExportMachineCACertsToFile(target string) error {
	if filesystem.FileExists(target) {
		return nil
	}

	var found []string
	var bundledCerts []byte
	for _, bundle := range CaBundles {
		for _, path := range bundle {
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
				cert, readErr := os.ReadFile(path)
				if readErr != nil {
					return errors.Join(ErrFailedToReadBundleFile, readErr)
				}
				bundledCerts = append(bundledCerts, cert...)
			}
		}
		if len(found) > 0 {
			break
		}
	}

	if len(bundledCerts) == 0 {
		return ErrNoCertBundleFound
	}

	err := os.MkdirAll(filepath.Dir(target), os.ModePerm)
	if err != nil {
		return errors.Join(ErrFailedToCreateCertDir, err)
	}
	err = os.WriteFile(target, bundledCerts, os.ModePerm)
	if err != nil {
		return errors.Join(ErrFailedToWriteCertFile, err)
	}

	log.Debug().Str("file", target).Msg("merged CA bundle file written")
	return nil
}

// EgressCABundleEnv points to the ca bundle (system roots + inspection ca) of the active egress filter, set for the
// process that runs under the filter. The file only exists while the filter inspects tls (type http endpoints).
const EgressCABundleEnv = "CID_EGRESS_CA_BUNDLE"

// EgressCABundleInContainer is where the egress ca bundle is mounted in action / executable containers
const EgressCABundleInContainer = "/cid/certs/egress-ca-bundle.crt"

// egressCABundleVariables point common clients to the mounted bundle (it contains the system roots, so it can replace
// the image's trust store). Java needs a truststore and is not covered.
var egressCABundleVariables = []string{
	"SSL_CERT_FILE",       // openssl (python ssl, ruby, .net on linux), go, rust (openssl-probe / rustls-native-certs), uv
	"NODE_EXTRA_CA_CERTS", // node / npm / yarn / pnpm
	"REQUESTS_CA_BUNDLE",  // python requests / pip
	"PIP_CERT",            // pip
	"CURL_CA_BUNDLE",      // curl
	"GIT_SSL_CAINFO",      // git
	"CARGO_HTTP_CAINFO",   // cargo
}

// MountEgressCABundle mounts the ca bundle of the active egress filter into the container and points common clients to
// it, so tls connections the filter inspects are trusted. No-op if the filter doesn't inspect tls.
func MountEgressCABundle(c *containerruntime.Container) {
	path := os.Getenv(EgressCABundleEnv)
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		return
	}

	c.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    CertMountSource(path),
		Target:    EgressCABundleInContainer,
		Mode:      containerruntime.ReadMode,
	})
	for _, name := range egressCABundleVariables {
		c.AddEnvironmentVariable(name, EgressCABundleInContainer)
	}
}

// CertMountSource returns the path to mount a cert file from, copied into CID_CERT_MOUNT_DIR if set (workaround for
// some dind setups, where only that directory is shared with the container runtime)
func CertMountSource(certPath string) string {
	customCertDir := os.Getenv("CID_CERT_MOUNT_DIR")
	if customCertDir == "" {
		return certPath
	}
	_ = os.MkdirAll(customCertDir, os.ModePerm)
	certDestinationFile := filepath.Join(customCertDir, filepath.Base(certPath))
	_ = filesystem.CopyFile(certPath, certDestinationFile)
	return certDestinationFile
}
