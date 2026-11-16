package egress

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParentPID(t *testing.T) {
	ppid, ok := parseParentPID("123 (my (weird) proc) S 45 123 123 0 -1")
	assert.True(t, ok)
	assert.Equal(t, 45, ppid)

	_, ok = parseParentPID("garbage")
	assert.False(t, ok)
}

func TestKillProcessTree(t *testing.T) {
	// sh starts a background sleep (grandchild of the test) and waits for it
	cmd := exec.Command("sh", "-c", "sleep 60 & wait")
	require.NoError(t, cmd.Start())

	var tree []int
	require.Eventually(t, func() bool {
		tree = processTree(cmd.Process.Pid)
		return len(tree) == 2
	}, 5*time.Second, 10*time.Millisecond, "sleep was not started")

	killProcessTree(cmd.Process)
	_ = cmd.Wait()

	assert.Eventually(t, func() bool {
		return len(processTree(cmd.Process.Pid)) == 1 && !processExists(tree[1])
	}, 5*time.Second, 10*time.Millisecond, "grandchild still running")
}

// processExists reports whether pid runs (zombies count as gone)
func processExists(pid int) bool {
	stat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false
	}
	end := strings.LastIndexByte(string(stat), ')')
	fields := strings.Fields(string(stat)[end+1:])
	return len(fields) > 0 && fields[0] != "Z"
}
