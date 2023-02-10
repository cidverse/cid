package command

import (
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/core/util"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
)

func ApplyProxyConfiguration(containerExec *containerruntime.Container) {
	// proxy
	containerExec.AddEnvironmentVariable("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	containerExec.AddEnvironmentVariable("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	containerExec.AddEnvironmentVariable("NO_PROXY", os.Getenv("NO_PROXY"))
	containerExec.AddEnvironmentVariable("http_proxy", os.Getenv("HTTP_PROXY"))
	containerExec.AddEnvironmentVariable("https_proxy", os.Getenv("HTTPS_PROXY"))
	containerExec.AddEnvironmentVariable("no_proxy", os.Getenv("NO_PROXY"))
}

// GetCertFileByType returns the cert file by type (ca-bundle, java-keystore)
func GetCertFileByType(certFileType string) string {
	var files []string

	if certFileType == "ca-bundle" {
		files = append(files, filepath.Join(util.GetUserConfigDirectory(), "certs", "ca-bundle.crt"))
		files = append(files, "/etc/pki/tls/certs/ca-bundle.crt")
		files = append(files, "/etc/ssl/certs/ca-certificates.crt")
	} else if certFileType == "java-keystore" {
		files = append(files, filepath.Join(util.GetUserConfigDirectory(), "certs", "keystore.jks"))
	}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}

	return ""
}

func ApplyCertMount(containerExec *containerruntime.Container, certFile string, containerCertFile string) {
	if certFile != "" {
		customCertDir := os.Getenv("CID_CERT_MOUNT_DIR")
		if customCertDir != "" {
			// Copy certFile into customCertDir
			_ = os.MkdirAll(customCertDir, os.ModePerm)
			certDestinationFile := filepath.Join(customCertDir, filepath.Base(certFile))
			_ = filesystem.CopyFile(certFile, certDestinationFile)

			// Overwrite certFile with new path of file in customCertDir
			certFile = certDestinationFile
		}

		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    certFile,
			Target:    containerCertFile,
			Mode:      containerruntime.ReadMode,
		})
	}
}
