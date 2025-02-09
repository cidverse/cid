package candidate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ContainerCandidate is used for the execution using container images
type ContainerCandidate struct {
	BaseCandidate
	Image      string `yaml:"package,omitempty"`
	ImageCache []ContainerCache
	Mounts     []ContainerMount
	Security   ContainerSecurity
	Entrypoint *string
	Certs      []ContainerCerts `yaml:"certs,omitempty"`
}

func (c ContainerCandidate) GetUri() string {
	return fmt.Sprintf("oci://%s", c.Image)
}

func (c ContainerCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running ContainerCandidate %s with args %v", c.Image, opts.Args)

	var stdoutWriter io.Writer
	var stderrWriter io.Writer
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if opts.CaptureOutput {
		stdoutWriter = redact.NewProtectedWriter(nil, &stdoutBuffer, &sync.Mutex{}, nil)
		stderrWriter = redact.NewProtectedWriter(nil, &stderrBuffer, &sync.Mutex{}, nil)
	} else {
		stdoutWriter = redact.NewProtectedWriter(os.Stdout, nil, &sync.Mutex{}, nil)
		stderrWriter = redact.NewProtectedWriter(os.Stderr, nil, &sync.Mutex{}, nil)
	}

	// overwrite binary for alias use-case
	commandArgs := append([]string{c.GetName()}, opts.Args...)
	containerExec := containerruntime.Container{
		Image:            c.Image,
		WorkingDirectory: ci.ToUnixPath(opts.WorkDir),
		Entrypoint:       c.Entrypoint,
		Command:          ci.ToUnixPathArgs(strings.Join(commandArgs, " ")),
		User:             util.GetContainerUser(),
	}

	// interactive?
	if opts.Stdin != nil {
		containerExec.Interactive = true
		containerExec.TTY = true
	}

	// security
	if c.Security.Privileged {
		containerExec.Privileged = true
	}
	containerExec.Capabilities = append(containerExec.Capabilities, c.Security.Capabilities...)

	// mounts
	containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: opts.RootDir, Target: ci.ToUnixPath(opts.RootDir)})
	if opts.TempDir != "" {
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: opts.TempDir, Target: opts.TempDir})
	}
	for _, mount := range c.Mounts {
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: mount.Src, Target: mount.Dest})
	}

	// add env + sort by key
	sortedEnvKeys := lo.Keys(opts.Env)
	sort.Strings(sortedEnvKeys)
	for _, key := range sortedEnvKeys {
		containerExec.AddEnvironmentVariable(key, opts.Env[key])
	}

	// cache
	for _, c := range c.ImageCache {
		dir := filepath.Join(util.CIDStateDir(), "cache-"+c.ID)
		_ = os.MkdirAll(dir, 0775)
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: dir, Target: c.ContainerPath})
	}

	// ports
	/*
		for _, port := range c.Ports {
			if network.IsFreePort(port) {
				containerExec.ContainerPorts = append(containerExec.ContainerPorts, containerruntime.ContainerPort{Source: port, Target: port})
			} else {
				freePort, _ := network.FreePort()
				containerExec.ContainerPorts = append(containerExec.ContainerPorts, containerruntime.ContainerPort{Source: freePort, Target: port})
			}
		}
	*/

	// enterprise (proxy, ca-certs)
	containerExec.AutoProxyConfiguration()
	for _, cert := range c.Certs {
		certPath, certErr := util.GetCertFileByType(cert.Type)
		if certErr != nil {
			return "", "", errors.New("failed to get cert file: " + certErr.Error())
		}

		// copy files into a custom directory if CID_CERT_MOUNT_DIR is set, workaround for some dind setups
		customCertDir := os.Getenv("CID_CERT_MOUNT_DIR")
		if customCertDir != "" {
			_ = os.MkdirAll(customCertDir, os.ModePerm)
			certDestinationFile := filepath.Join(customCertDir, filepath.Base(certPath))
			_ = filesystem.CopyFile(certPath, certDestinationFile)

			certPath = certDestinationFile
		}

		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    certPath,
			Target:    cert.ContainerPath,
			Mode:      containerruntime.ReadMode,
		})
	}

	containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if containerCmdErr != nil {
		return "", "", containerCmdErr
	}

	cmd, err := shellcommand.PrepareCommand(containerCmd, runtime.GOOS, "bash", true, nil, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
