package egress

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

// killProcessTree stops the process and all its descendants (SIGSTOP, so nothing can fork while the tree is collected)
// and then kills them. Processes that detached themselves (e.g. podman's conmon, reparented to init) are not part of
// the tree; in ci they end with the job container.
func killProcessTree(root *os.Process) {
	killed := map[int]bool{}
	// two passes: processes forked while the first pass was stopping their parent are caught in the second
	for range 2 {
		for _, pid := range processTree(root.Pid) {
			if !killed[pid] {
				_ = syscall.Kill(pid, syscall.SIGSTOP)
				killed[pid] = true
			}
		}
	}
	for pid := range killed {
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && pid == root.Pid {
			log.Warn().Err(err).Int("pid", pid).Msg("failed to kill process")
		}
	}
	if len(killed) == 0 {
		_ = root.Kill()
	}
}

// processTree returns root and all its descendants, from /proc
func processTree(root int) []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return []int{root}
	}
	children := map[int][]int{}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		stat, err := os.ReadFile("/proc/" + e.Name() + "/stat")
		if err != nil {
			continue // exited
		}
		if ppid, ok := parseParentPID(string(stat)); ok {
			children[ppid] = append(children[ppid], pid)
		}
	}

	tree := []int{root}
	for i := 0; i < len(tree); i++ {
		tree = append(tree, children[tree[i]]...)
	}
	return tree
}

// parseParentPID reads the ppid from /proc/<pid>/stat: `pid (comm) state ppid ...`, comm may contain spaces and parentheses
func parseParentPID(stat string) (int, bool) {
	end := strings.LastIndexByte(stat, ')')
	if end < 0 {
		return 0, false
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 2 {
		return 0, false
	}
	ppid, err := strconv.Atoi(fields[1])
	return ppid, err == nil
}
