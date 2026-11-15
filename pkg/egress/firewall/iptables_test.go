package firewall

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingRunner struct {
	commands []string
	fail     func(cmd string) bool
}

func (r *recordingRunner) run(name string, args ...string) error {
	cmd := name + " " + strings.Join(args, " ")
	if r.fail != nil && r.fail(cmd) {
		return fmt.Errorf("failed")
	}
	r.commands = append(r.commands, cmd)
	return nil
}

// teardownFails simulates a clean system: removing leftovers of a previous run fails
func teardownFails(cmd string) bool {
	return strings.Contains(cmd, " -D ") || strings.Contains(cmd, " -F ") || strings.Contains(cmd, " -X ")
}

func TestIPTablesSetupBlockMode(t *testing.T) {
	r := &recordingRunner{fail: teardownFails}
	fw := &IPTables{Mode: policy.ModeBlock, Mark: 0xc1d0, NFLogGroup: 1043, Run: r.run}

	static := []policy.Endpoint{
		{Prefix: netip.MustParsePrefix("10.0.0.0/8"), Port: 443},
		{Prefix: netip.MustParsePrefix("2001:db8::/32")},
	}
	require.NoError(t, fw.Setup(static, testRedirect))

	assert.Equal(t, []string{
		"iptables -w -N CID-EGRESS",
		"iptables -w -A CID-EGRESS -o lo -j RETURN",
		"iptables -w -A CID-EGRESS -d 127.0.0.0/8 -j RETURN",
		"iptables -w -A CID-EGRESS -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN",
		"iptables -w -A CID-EGRESS -m mark --mark 0xc1d0 -j RETURN",
		"iptables -w -A CID-EGRESS -d 10.0.0.0/8 -p tcp --dport 443 -j RETURN",
		"iptables -w -A CID-EGRESS -d 10.0.0.0/8 -p udp --dport 443 -j RETURN",
		"iptables -w -A CID-EGRESS -j NFLOG --nflog-group 1043 --nflog-prefix cid-egress",
		"iptables -w -A CID-EGRESS -p tcp -j REJECT --reject-with tcp-reset",
		"iptables -w -A CID-EGRESS -j REJECT",
		"iptables -w -I OUTPUT 1 -j CID-EGRESS",
		"iptables -w -I FORWARD 1 -j CID-EGRESS",
		"iptables -w -t nat -N CID-EGRESS",
		"iptables -w -t nat -A CID-EGRESS -m mark --mark 0xc1d0 -j RETURN",
		"iptables -w -t nat -A CID-EGRESS -d 198.18.0.0/15 -p tcp -j REDIRECT --to-ports 15001",
		"iptables -w -t nat -I OUTPUT 1 -j CID-EGRESS",
		"ip6tables -w -N CID-EGRESS",
		"ip6tables -w -A CID-EGRESS -o lo -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ::1/128 -j RETURN",
		"ip6tables -w -A CID-EGRESS -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN",
		"ip6tables -w -A CID-EGRESS -d fe80::/10 -p ipv6-icmp --icmpv6-type 133 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d fe80::/10 -p ipv6-icmp --icmpv6-type 135 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d fe80::/10 -p ipv6-icmp --icmpv6-type 136 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d fe80::/10 -p ipv6-icmp --icmpv6-type 131 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d fe80::/10 -p ipv6-icmp --icmpv6-type 143 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ff02::/16 -p ipv6-icmp --icmpv6-type 133 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ff02::/16 -p ipv6-icmp --icmpv6-type 135 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ff02::/16 -p ipv6-icmp --icmpv6-type 136 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ff02::/16 -p ipv6-icmp --icmpv6-type 131 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d ff02::/16 -p ipv6-icmp --icmpv6-type 143 -j RETURN",
		"ip6tables -w -A CID-EGRESS -m mark --mark 0xc1d0 -j RETURN",
		"ip6tables -w -A CID-EGRESS -d 2001:db8::/32 -j RETURN",
		"ip6tables -w -A CID-EGRESS -j NFLOG --nflog-group 1043 --nflog-prefix cid-egress",
		"ip6tables -w -A CID-EGRESS -p tcp -j REJECT --reject-with tcp-reset",
		"ip6tables -w -A CID-EGRESS -j REJECT",
		"ip6tables -w -I OUTPUT 1 -j CID-EGRESS",
		"ip6tables -w -I FORWARD 1 -j CID-EGRESS",
	}, r.commands, "the ipv4 redirect is only added to iptables")
}

func TestIPTablesAuditModeHasNoReject(t *testing.T) {
	r := &recordingRunner{}
	fw := &IPTables{Mode: policy.ModeAudit, NFLogGroup: 1043, Run: r.run}
	require.NoError(t, fw.Setup(nil, testRedirect))
	for _, cmd := range r.commands {
		assert.NotContains(t, cmd, "REJECT")
	}
}

func TestIPTablesClose(t *testing.T) {
	r := &recordingRunner{fail: func(cmd string) bool { return teardownFails(cmd) || strings.HasPrefix(cmd, "ip6tables") }}
	fw := &IPTables{Mode: policy.ModeBlock, Run: r.run}
	require.NoError(t, fw.Setup(nil, testRedirect), "missing ip6tables is not fatal")

	r.commands = nil
	r.fail = nil
	require.NoError(t, fw.Close())
	for _, cmd := range r.commands {
		assert.True(t, strings.HasPrefix(cmd, "iptables "), "only the families that were set up are removed: %s", cmd)
	}
	assert.Contains(t, r.commands, "iptables -w -t filter -X CID-EGRESS")
	assert.Contains(t, r.commands, "iptables -w -t nat -X CID-EGRESS")
}

func TestIPTablesRequiresRunner(t *testing.T) {
	fw := &IPTables{Mode: policy.ModeBlock}
	assert.Error(t, fw.Setup(nil, testRedirect))
	assert.NoError(t, fw.Close())
}

func TestIPTablesMissingIPv4IsFatal(t *testing.T) {
	r := &recordingRunner{fail: func(cmd string) bool { return strings.HasPrefix(cmd, "iptables -w -N") }}
	fw := &IPTables{Mode: policy.ModeBlock, Run: r.run}
	assert.Error(t, fw.Setup(nil, testRedirect))
}
