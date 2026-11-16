package hardening

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

// capabilityValues maps DroppedCapabilities to their numbers
var capabilityValues = map[string]uintptr{
	"NET_ADMIN":       unix.CAP_NET_ADMIN,
	"NET_RAW":         unix.CAP_NET_RAW,
	"SYS_ADMIN":       unix.CAP_SYS_ADMIN,
	"SYS_MODULE":      unix.CAP_SYS_MODULE,
	"SYS_RAWIO":       unix.CAP_SYS_RAWIO,
	"DAC_READ_SEARCH": unix.CAP_DAC_READ_SEARCH,
	"SYS_PTRACE":      unix.CAP_SYS_PTRACE,
	"BPF":             unix.CAP_BPF,
	"PERFMON":         unix.CAP_PERFMON,
}

// WrapCommand makes a prepared local command start through the exec shim if the egress filter is active, no-op
// otherwise. Must be called before cmd.Start.
func WrapCommand(cmd *exec.Cmd) error {
	mode := ActiveMode(os.Getenv)
	if mode == "" || cmd.Err != nil {
		return nil // not under the filter / lookup failed, Start reports the error
	}
	self, err := os.Executable()
	if err != nil {
		if mode == policy.ModeBlock {
			return fmt.Errorf("egress filter: failed to determine the cid executable for the exec shim: %w", err)
		}
		log.Warn().Err(err).Msg("egress filter: failed to determine the cid executable, running the command with full privileges (audit mode)")
		return nil
	}
	cmd.Args = append([]string{self, ExecShimArg, string(mode), cmd.Path, "--"}, cmd.Args...)
	cmd.Path = self
	return nil
}

// RunExecShim drops the privileges and executes the command, never returns. args are the arguments after ExecShimArg:
// `<mode> <path> -- <argv...>`. In block mode the command is not executed if the privileges can't be dropped.
func RunExecShim(args []string) {
	if len(args) < 4 || args[2] != "--" {
		fmt.Fprintf(os.Stderr, "cid egress: invalid exec shim arguments %q\n", args)
		os.Exit(126)
	}
	mode, path, argv := policy.Mode(args[0]), args[1], args[3:]

	// capabilities and no_new_privs are per thread, drop them on the thread that calls execve
	runtime.LockOSThread()
	if err := dropPrivileges(); err != nil {
		if mode != policy.ModeAudit {
			fmt.Fprintf(os.Stderr, "cid egress: failed to drop privileges, %s was not executed: %v\n", path, err)
			os.Exit(126)
		}
		fmt.Fprintf(os.Stderr, "cid egress: warning: failed to drop privileges, running %s with full privileges (audit mode): %v\n", path, err)
	}

	err := syscall.Exec(path, argv, os.Environ())
	fmt.Fprintf(os.Stderr, "cid egress: failed to execute %s: %v\n", path, err)
	os.Exit(127)
}

// dropPrivileges removes DroppedCapabilities from the bounding, ambient, inheritable, permitted and effective sets and
// sets no_new_privs. For root the capabilities after execve are (inheritable | bounding | ambient), all three must be
// cleared.
func dropPrivileges() error {
	for _, name := range DroppedCapabilities {
		value := capabilityValues[name]
		if err := unix.Prctl(unix.PR_CAPBSET_DROP, value, 0, 0, 0); err != nil {
			if errors.Is(err, unix.EINVAL) {
				continue // unknown to this kernel (BPF / PERFMON before 5.8), can't be granted either
			}
			return fmt.Errorf("drop CAP_%s from the bounding set: %w", name, err)
		}
		if set, err := unix.PrctlRetInt(unix.PR_CAPBSET_READ, value, 0, 0, 0); err != nil || set != 0 {
			return fmt.Errorf("CAP_%s is still in the bounding set", name)
		}
	}

	if err := unix.Prctl(unix.PR_CAP_AMBIENT, unix.PR_CAP_AMBIENT_CLEAR_ALL, 0, 0, 0); err != nil && !errors.Is(err, unix.EINVAL) {
		return fmt.Errorf("clear the ambient capabilities: %w", err)
	}

	hdr := unix.CapUserHeader{Version: unix.LINUX_CAPABILITY_VERSION_3}
	var data [2]unix.CapUserData // version 3 uses two 32 bit words
	if err := unix.Capget(&hdr, &data[0]); err != nil {
		return fmt.Errorf("read the capabilities: %w", err)
	}
	for _, name := range DroppedCapabilities {
		value := capabilityValues[name]
		bit := uint32(1) << (value % 32)
		word := &data[value/32]
		word.Inheritable &^= bit
		word.Permitted &^= bit
		word.Effective &^= bit
	}
	if err := unix.Capset(&hdr, &data[0]); err != nil {
		return fmt.Errorf("drop the capabilities: %w", err)
	}

	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("set no_new_privs: %w", err)
	}
	return nil
}
