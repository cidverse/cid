package executable

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/cidverse/cid/pkg/common/shellcommand"
	"github.com/cidverse/cid/pkg/egress/hardening"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/cidverse/cidverseutils/containerruntime"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
)

// ContainerCandidate is used for the execution using container images
type ContainerCandidate struct {
	BaseCandidate
	Image      string `yaml:"package,omitempty"`
	Digest     string `yaml:"digest,omitempty"` // Digest is the manifest digest the image tag was resolved to
	ImageCache []ContainerCache
	Mounts     []ContainerMount
	Security   ContainerSecurity
	Entrypoint *string
	Certs      []ContainerCerts `yaml:"certs,omitempty"`
}

// Reference returns the digest-pinned image reference, falling back to the tag if no digest is known.
func (c ContainerCandidate) Reference() string {
	return PinnedReference(c.Image, c.Digest)
}

// NamedReference returns the human readable tag reference of the image, empty for digest-only images.
func (c ContainerCandidate) NamedReference() string {
	return NamedReference(c.Image)
}

func (c ContainerCandidate) GetUri() string {
	return fmt.Sprintf("oci://%s", c.Reference())
}

func (c ContainerCandidate) GetNamedReference() string {
	return c.NamedReference()
}

func (c ContainerCandidate) Run(opts RunParameters) (string, string, error) {
	log.Debug().Msgf("Running ContainerCandidate %s with args %v", c.Reference(), opts.Args)

	if c.Digest == "" {
		log.Warn().Str("image", c.Image).Msg("running container image without a pinned digest, the tag may have been moved since discovery - run 'cid executables update' to refresh the candidate cache")
	}

	var stdoutBuffer, stderrBuffer bytes.Buffer
	var stdoutWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), &stdoutBuffer)
	var stderrWriter = io.MultiWriter(redact.NewProtectedWriter(nil, os.Stderr, &sync.Mutex{}, nil), &stderrBuffer)
	if opts.HideStdOut {
		stdoutWriter = &stdoutBuffer
	}
	if opts.HideStdErr {
		stderrWriter = &stderrBuffer
	}

	// overwrite binary for alias use-case
	containerUser := util.GetContainerUser()
	containerExec := containerruntime.Container{
		Image:            c.Reference(),
		WorkingDirectory: ci.ToUnixPath(opts.WorkDir),
		Entrypoint:       c.Entrypoint,
		Command:          ci.ToUnixPathArgs(strings.Join(opts.Args, " ")),
		User:             containerUser,
	}

	// interactive?
	if opts.Stdin != nil {
		containerExec.Interactive = true
		containerExec.TTY = true
	}

	// security: privileged containers are refused, dangerous capabilities dropped (see hardening.ContainerArgs)
	containerExec.Capabilities = append(containerExec.Capabilities, c.Security.Capabilities...)
	hardeningArgs, err := hardening.ContainerArgs(c.Reference(), c.Security.Privileged, containerExec.Capabilities)
	if err != nil {
		return "", "", err
	}
	containerExec.UserArgs = hardening.AppendArgs(containerExec.UserArgs, hardeningArgs)

	// mounts
	containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: opts.RootDir, Target: ci.ToUnixPath(opts.RootDir)})
	if opts.TempDir != "" {
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: opts.TempDir, Target: ci.ToUnixPath(opts.TempDir)})
	}
	for _, mount := range c.Mounts {
		containerExec.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: mount.Src, Target: mount.Dest})
	}

	// add env + sort by key
	sortedEnvKeys := slices.Collect(maps.Keys(opts.Env))
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
	util.MountEgressCABundle(&containerExec)
	for _, cert := range c.Certs {
		certPath, certErr := util.GetCertFileByType(cert.Type)
		if certErr != nil {
			return "", "", errors.New("failed to get cert file: " + certErr.Error())
		}

		containerExec.AddVolume(containerruntime.ContainerMount{
			MountType: "directory",
			Source:    util.CertMountSource(certPath),
			Target:    cert.ContainerPath,
			Mode:      containerruntime.ReadMode,
		})
	}

	containerCmd, containerCmdErr := containerExec.GetRunCommand(containerExec.DetectRuntime())
	if containerCmdErr != nil {
		return "", "", containerCmdErr
	}

	cmd, err := shellcommand.PrepareCommand(containerCmd, runtime.GOOS, "", true, map[string]string{"PODMAN_IGNORE_CGROUPSV1_WARNING": "true"}, opts.WorkDir, opts.Stdin, stdoutWriter, stderrWriter)
	if err != nil {
		return "", "", err
	}

	err = cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("error running command: %w", err)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
