package command

import (
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/core/util"
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
		filepath.Join(util.GetUserConfigDirectory(), "ca-extra.crt"),
		"/etc/pki/tls/certs/ca-extra.crt",
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
		Target:    "/etc/pki/tls/certs/ca-extra.crt",
		Mode:      containerruntime.ReadMode,
	})
}
