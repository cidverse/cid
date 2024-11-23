package command

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
)

func ApplyProxyConfiguration(containerExec *containerruntime.Container) {
	// proxy
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")
	noProxy := os.Getenv("NO_PROXY")

	if httpProxy != "" {
		containerExec.AddEnvironmentVariable("HTTP_PROXY", httpProxy)
		containerExec.AddEnvironmentVariable("http_proxy", httpProxy)
	}
	if httpsProxy != "" {
		containerExec.AddEnvironmentVariable("HTTPS_PROXY", httpsProxy)
		containerExec.AddEnvironmentVariable("https_proxy", httpsProxy)
	}
	if noProxy != "" {
		containerExec.AddEnvironmentVariable("NO_PROXY", noProxy)
		containerExec.AddEnvironmentVariable("no_proxy", noProxy)
	}

	// jvm
	var javaProxyOpts []string
	if httpProxy != "" {
		proxyURL, err := url.Parse(httpProxy)
		if err == nil {
			javaProxyOpts = append(javaProxyOpts, "-Dhttp.proxyHost="+proxyURL.Hostname())
			javaProxyOpts = append(javaProxyOpts, "-Dhttp.proxyPort="+proxyURL.Port())
			javaProxyOpts = append(javaProxyOpts, "-Dhttp.nonProxyHosts="+ConvertNoProxyForJava(noProxy))
		}
	}
	if httpsProxy != "" {
		proxyURL, err := url.Parse(httpsProxy)
		if err == nil {
			javaProxyOpts = append(javaProxyOpts, "-Dhttps.proxyHost="+proxyURL.Hostname())
			javaProxyOpts = append(javaProxyOpts, "-Dhttps.proxyPort="+proxyURL.Port())
			javaProxyOpts = append(javaProxyOpts, "-Dhttps.nonProxyHosts="+ConvertNoProxyForJava(noProxy))
		}
	}
	if len(javaProxyOpts) > 0 {
		containerExec.AddEnvironmentVariable("CID_PROXY_JVM", strings.Join(javaProxyOpts, " "))
	}
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

// ReplaceCommandPlaceholders replaces env placeholders in a command
func ReplaceCommandPlaceholders(input string, env map[string]string) string {
	// timestamp
	input = strings.ReplaceAll(input, "{TIMESTAMP_RFC3339}", time.Now().Format(time.RFC3339))

	// env
	for k, v := range env {
		input = strings.ReplaceAll(input, "{"+k+"}", v)
	}

	return input
}

func ConvertNoProxyForJava(input string) string {
	return strings.ReplaceAll(input, ",", "|")
}
