// Package hardening removes dangerous privileges (container escapes, disabling / bypassing the egress filter) from the
// processes cid starts (see ../README.md, "Hardening"):
//
//   - containers always get `--cap-drop` for DroppedCapabilities and `--security-opt no-new-privileges`, privileged
//     containers and containers that request one of these capabilities are never started
//   - local commands under the egress filter are started through an exec shim (cid re-executed with ExecShimArg) that
//     removes the capabilities from the bounding, inheritable and ambient sets and sets no_new_privs before it executes
//     the command
//
// cid itself keeps its privileges: it manages the filter and starts podman.
package hardening

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/egress/policy"
)

const (
	// EnvActive is set to the filter mode for the cid process that runs under the egress filter
	EnvActive = "CID_EGRESS_FILTER_ACTIVE"
	// ExecShimArg is the first argument of the exec shim: `cid __cid-egress-exec <mode> <path> -- <argv...>`
	ExecShimArg = "__cid-egress-exec"
)

// DroppedCapabilities are removed from every container and from local commands under the egress filter:
//   - NET_ADMIN, NET_RAW: change the firewall, set the egress filter's socket mark (SO_MARK) to bypass it
//   - SYS_ADMIN, SYS_MODULE, SYS_RAWIO, DAC_READ_SEARCH: mounts / namespaces, kernel modules, raw devices,
//     open_by_handle_at, the usual ways out of the job container
//   - SYS_PTRACE: attach to other processes, e.g. the cid process that runs the egress filter
//   - BPF, PERFMON: load bpf programs that could rewrite packets or sockets
var DroppedCapabilities = []string{
	"NET_ADMIN",
	"NET_RAW",
	"SYS_ADMIN",
	"SYS_MODULE",
	"SYS_RAWIO",
	"DAC_READ_SEARCH",
	"SYS_PTRACE",
	"BPF",
	"PERFMON",
}

// ActiveMode returns the mode of the egress filter the current process runs under, empty if it doesn't run under the
// filter. Unknown values are treated as block mode.
func ActiveMode(getenv func(string) string) policy.Mode {
	value := getenv(EnvActive)
	if value == "" {
		return ""
	}
	mode, err := policy.ParseMode(value)
	if err != nil {
		return policy.ModeBlock
	}
	return mode
}

// ContainerArgs returns the additional container run arguments (podman / docker) for every container cid starts: drop
// DroppedCapabilities and set no-new-privileges. Privileged containers and containers that request one of
// DroppedCapabilities are an error, with or without the egress filter.
func ContainerArgs(image string, privileged bool, capabilities []string) (string, error) {
	if privileged {
		return "", fmt.Errorf("container %s requests privileged mode, which is not supported", image)
	}

	var refused []string
	for _, c := range capabilities {
		name := normalizeCapability(c)
		if slices.Contains(DroppedCapabilities, name) && !slices.Contains(refused, name) {
			refused = append(refused, name)
		}
	}
	if len(refused) > 0 {
		return "", fmt.Errorf("container %s requests capabilities %v, which are not supported (container escapes, disabling the egress filter)", image, refused)
	}

	args := make([]string, 0, len(DroppedCapabilities)+1)
	for _, c := range DroppedCapabilities {
		args = append(args, "--cap-drop="+c)
	}
	args = append(args, "--security-opt=no-new-privileges")
	return strings.Join(args, " "), nil
}

// AppendArgs appends args to a space separated argument string
func AppendArgs(existing string, args string) string {
	switch {
	case args == "":
		return existing
	case existing == "":
		return args
	default:
		return existing + " " + args
	}
}

// normalizeCapability returns the capability name without CAP_ prefix, upper case
func normalizeCapability(c string) string {
	c = strings.ToUpper(strings.TrimSpace(c))
	return strings.TrimPrefix(c, "CAP_")
}
