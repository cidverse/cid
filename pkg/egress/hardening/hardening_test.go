package hardening

import (
	"testing"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envWith(mode string) func(string) string {
	return func(k string) string {
		if k == EnvActive {
			return mode
		}
		return ""
	}
}

func TestActiveMode(t *testing.T) {
	assert.Equal(t, policy.Mode(""), ActiveMode(envWith("")))
	assert.Equal(t, policy.ModeAudit, ActiveMode(envWith("audit")))
	assert.Equal(t, policy.ModeBlock, ActiveMode(envWith("block")))
	assert.Equal(t, policy.ModeBlock, ActiveMode(envWith("1")), "unknown values fail safe")
}

func TestContainerArgs(t *testing.T) {
	args, err := ContainerArgs("img", false, []string{"CHOWN"})
	require.NoError(t, err)
	assert.Equal(t, "--cap-drop=NET_ADMIN --cap-drop=NET_RAW --cap-drop=SYS_ADMIN --cap-drop=SYS_MODULE --cap-drop=SYS_RAWIO --cap-drop=DAC_READ_SEARCH --cap-drop=SYS_PTRACE --cap-drop=BPF --cap-drop=PERFMON --security-opt=no-new-privileges", args)

	_, err = ContainerArgs("img", true, nil)
	require.ErrorContains(t, err, "requests privileged mode, which is not supported")

	_, err = ContainerArgs("img", false, []string{"cap_net_admin", "NET_ADMIN", "CHOWN"})
	require.ErrorContains(t, err, "requests capabilities [NET_ADMIN], which are not supported")
}

func TestAppendArgs(t *testing.T) {
	assert.Equal(t, "--net host", AppendArgs("--net host", ""))
	assert.Equal(t, "--cap-drop=X", AppendArgs("", "--cap-drop=X"))
	assert.Equal(t, "--net host --cap-drop=X", AppendArgs("--net host", "--cap-drop=X"))
}
