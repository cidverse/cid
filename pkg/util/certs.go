package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
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
