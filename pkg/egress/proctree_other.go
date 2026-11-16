//go:build !linux

package egress

import "os"

// killProcessTree kills the process, the egress filter only runs on linux
func killProcessTree(root *os.Process) {
	_ = root.Kill()
}
