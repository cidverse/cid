package hardening

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const testShimEnv = "CID_HARDENING_TEST_SHIM"

// TestMain lets the test binary act as the exec shim (like cid's main)
func TestMain(m *testing.M) {
	if os.Getenv(testShimEnv) == "1" {
		RunExecShim(os.Args[1:])
	}
	os.Exit(m.Run())
}

func TestWrapCommand(t *testing.T) {
	cmd := exec.Command("sh", "-c", "true")
	original := cmd.Path

	t.Setenv(EnvActive, "")
	require.NoError(t, WrapCommand(cmd))
	assert.Equal(t, original, cmd.Path, "unchanged without the filter")

	t.Setenv(EnvActive, "block")
	require.NoError(t, WrapCommand(cmd))
	self, err := os.Executable()
	require.NoError(t, err)
	assert.Equal(t, self, cmd.Path)
	assert.Equal(t, []string{self, ExecShimArg, "block", original, "--", "sh", "-c", "true"}, cmd.Args)
}

// runShim executes `sh -c script` through the shim (the test binary) and returns its output and exit code
func runShim(t *testing.T, mode string, script string) (string, int) {
	t.Helper()
	sh, err := exec.LookPath("sh")
	require.NoError(t, err)
	self, err := os.Executable()
	require.NoError(t, err)

	cmd := exec.Command(self, mode, sh, "--", "sh", "-c", script)
	cmd.Env = append(os.Environ(), testShimEnv+"=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return out.String(), exitErr.ExitCode()
	}
	require.NoError(t, err)
	return out.String(), 0
}

func TestExecShim(t *testing.T) {
	// shell builtins only (no grep / cat), minimal ci images don't ship them; the shell reads its own status
	out, code := runShim(t, "block", `while IFS= read -r line; do echo "$line"; done < /proc/self/status`)

	if os.Geteuid() != 0 {
		// no CAP_SETPCAP: block mode must not run the command
		assert.Equal(t, 126, code, out)
		assert.Contains(t, out, "failed to drop privileges")

		out, code = runShim(t, "audit", "echo ran")
		assert.Equal(t, 0, code, out)
		assert.Contains(t, out, "ran")
		return
	}

	require.Equal(t, 0, code, out)
	status := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		k, v, _ := strings.Cut(line, ":")
		status[k] = strings.TrimSpace(v)
	}
	assert.Equal(t, "1", status["NoNewPrivs"])
	for _, set := range []string{"CapBnd", "CapInh", "CapAmb"} {
		value, err := strconv.ParseUint(status[set], 16, 64)
		require.NoError(t, err, set)
		for _, name := range DroppedCapabilities {
			assert.Zero(t, value&(1<<capabilityValues[name]), "%s still contains CAP_%s", set, name)
		}
	}
	assert.NotZero(t, mustBndHas(t, status["CapBnd"], unix.CAP_CHOWN), "other capabilities are kept")
}

func mustBndHas(t *testing.T, hex string, capability uint) uint64 {
	t.Helper()
	value, err := strconv.ParseUint(hex, 16, 64)
	require.NoError(t, err)
	return value & (1 << capability)
}
