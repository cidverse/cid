package firewall

import (
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRedirect = Redirect{Prefix: netip.MustParsePrefix("198.18.0.0/15"), Port: 15001}

type recordingScriptRunner struct {
	scripts []string
	fail    func(script string) bool
}

func (r *recordingScriptRunner) run(script string) error {
	if r.fail != nil && r.fail(script) {
		return errors.New("failed")
	}
	r.scripts = append(r.scripts, script)
	return nil
}

func TestNFTablesSetupBlockMode(t *testing.T) {
	r := &recordingScriptRunner{}
	fw := &NFTables{Mode: policy.ModeBlock, Mark: 0xc1d0, NFLogGroup: 1043, Run: r.run}

	static := []policy.Endpoint{
		{Prefix: netip.MustParsePrefix("10.0.0.0/8"), Port: 443},
		{Prefix: netip.MustParsePrefix("2001:db8::/32")},
	}
	require.NoError(t, fw.Setup(static, testRedirect))
	require.Len(t, r.scripts, 1)
	assert.Equal(t, `table inet cid_egress
delete table inet cid_egress
table inet cid_egress {
	chain egress {
		oifname "lo" accept
		ip daddr 127.0.0.0/8 accept
		ip6 daddr ::1/128 accept
		ct state established,related accept
		ip6 daddr { fe80::/10, ff02::/16 } icmpv6 type { nd-router-solicit, nd-neighbor-solicit, nd-neighbor-advert, mld-listener-report, mld2-listener-report } accept
		meta mark 0xc1d0 accept
		ip daddr 10.0.0.0/8 tcp dport 443 accept
		ip daddr 10.0.0.0/8 udp dport 443 accept
		ip6 daddr 2001:db8::/32 accept
		log prefix "cid-egress" group 1043
		meta l4proto tcp reject with tcp reset
		reject
	}
	chain output {
		type filter hook output priority 0; policy accept;
		jump egress
	}
	chain forward {
		type filter hook forward priority 0; policy accept;
		jump egress
	}
	chain nat_output {
		type nat hook output priority -100; policy accept;
		meta mark 0xc1d0 return
		ip daddr 198.18.0.0/15 meta l4proto tcp redirect to :15001
	}
}
`, r.scripts[0])

	r.scripts = nil
	require.NoError(t, fw.Close())
	assert.Equal(t, []string{"delete table inet cid_egress\n"}, r.scripts)
	require.NoError(t, fw.Close())
	assert.Len(t, r.scripts, 1, "closing twice is a no-op")
}

func TestNFTablesAuditModeHasNoReject(t *testing.T) {
	r := &recordingScriptRunner{}
	fw := &NFTables{Mode: policy.ModeAudit, NFLogGroup: 1043, Run: r.run}
	require.NoError(t, fw.Setup(nil, testRedirect))
	require.Len(t, r.scripts, 1)
	assert.NotContains(t, r.scripts[0], "reject")
	assert.Contains(t, r.scripts[0], "log prefix")
	assert.NotContains(t, r.scripts[0], "meta mark", "no mark configured")
}

func TestNFTablesFallsBackWithoutLog(t *testing.T) {
	r := &recordingScriptRunner{fail: func(script string) bool { return strings.Contains(script, "log prefix") }}
	fw := &NFTables{Mode: policy.ModeBlock, NFLogGroup: 1043, Run: r.run}
	require.NoError(t, fw.Setup(nil, testRedirect))
	require.Len(t, r.scripts, 1)
	assert.NotContains(t, r.scripts[0], "log prefix")
	assert.Contains(t, r.scripts[0], "reject")
}

func TestNFTablesRequiresRunner(t *testing.T) {
	fw := &NFTables{Mode: policy.ModeBlock}
	assert.Error(t, fw.Setup(nil, testRedirect))
	assert.NoError(t, fw.Close(), "nothing active, no-op")
}

func TestNFTablesSnapshot(t *testing.T) {
	var called []string
	fw := &NFTables{List: func(name string, args ...string) (string, error) {
		called = append([]string{name}, args...)
		return "ruleset", nil
	}}
	out, err := fw.Snapshot()
	require.NoError(t, err)
	assert.Equal(t, "ruleset", out)
	assert.Equal(t, []string{"nft", "-s", "list", "ruleset"}, called)

	_, err = (&NFTables{}).Snapshot()
	assert.Error(t, err)
}
