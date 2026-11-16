//go:build !linux

package hardening

import (
	"fmt"
	"os"
	"os/exec"
)

// WrapCommand is a no-op, the egress filter only runs on linux
func WrapCommand(_ *exec.Cmd) error {
	return nil
}

// RunExecShim is only supported on linux
func RunExecShim(_ []string) {
	fmt.Fprintln(os.Stderr, "cid egress: the exec shim is only supported on linux")
	os.Exit(126)
}
