package command

import (
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/core/global"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
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

func ApplyCertConfiguration(containerExec *containerruntime.Container) {
	files := []string{
		filepath.Join(global.GetUserConfigDirectory(), "ca-bundle.crt"),
		"/etc/pki/tls/certs/ca-bundle.crt",
		"/etc/ssl/certs/ca-certificates.crt",
	}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			ApplyCertMount(containerExec, file)
			break
		}
	}
}

func ApplyCertMount(containerExec *containerruntime.Container, bundleFile string) {
	containerExec.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    bundleFile,
		Target:    "/etc/pki/tls/certs/ca-bundle.crt",
		Mode:      containerruntime.ReadMode,
	})
	containerExec.AddVolume(containerruntime.ContainerMount{
		MountType: "directory",
		Source:    bundleFile,
		Target:    "/etc/ssl/certs/ca-certificates.crt",
		Mode:      containerruntime.ReadMode,
	})
}
