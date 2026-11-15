package firewall

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/rs/zerolog/log"
)

const nftTable = "inet cid_egress"

// ScriptRunner applies a nft script (`nft -f -`), replaceable for tests
type ScriptRunner func(script string) error

// ExecNFT applies the script with the nft binary on the host
func ExecNFT(script string) error {
	cmd := exec.Command("nft", "-f", "-")
	cmd.Stdin = strings.NewReader(script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nft -f -: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// NFTables manages a dedicated nftables table that only allows egress through the transparent proxy.
//
// All rules live in the table `inet cid_egress` (ipv4 + ipv6), other tables are never touched, and teardown is a single
// atomic `delete table`:
//
//	table inet cid_egress {
//		chain egress {
//			oifname "lo" accept
//			ip daddr 127.0.0.0/8 accept                   // connections redirected to the proxy, see loopbackPrefixes
//			ip6 daddr ::1 accept
//			ct state established,related accept
//			ip6 daddr { fe80::/10, ff02::/16 } icmpv6 type { nd-*, mld-* } accept  // ipv6 neighbor discovery
//			meta mark <mark> accept                       // the egress filter's own upstream connections
//			ip daddr <static-cidr> ... accept             // static ip / cidr entries
//			log prefix "cid-egress" group <group>         // report everything else
//			meta l4proto tcp reject with tcp reset        // block mode only
//			reject                                        // block mode only
//		}
//		chain output { type filter hook output priority 0; policy accept; jump egress; }
//		chain forward { type filter hook forward priority 0; policy accept; jump egress; }
//		chain nat_output {
//			type nat hook output priority -100; policy accept;
//			meta mark <mark> return
//			ip daddr 198.18.0.0/15 meta l4proto tcp redirect to :<proxy-port>
//		}
//	}
type NFTables struct {
	Mode       policy.Mode
	Mark       uint32
	NFLogGroup uint16       // 0 = disabled
	Run        ScriptRunner // required, use ExecNFT to modify the host firewall (no default, so tests can't touch the host by accident)

	active bool
}

func (f *NFTables) Setup(static []policy.Endpoint, redirect Redirect) error {
	if f.Run == nil {
		return errors.New("nftables: no script runner configured")
	}

	err := f.Run(f.setupScript(static, redirect, f.NFLogGroup != 0))
	if err != nil && f.NFLogGroup != 0 {
		log.Warn().Err(err).Msg("failed to set up nftables egress table with nflog, retrying without, blocked connections will not be reported")
		err = f.Run(f.setupScript(static, redirect, false))
	}
	if err != nil {
		return fmt.Errorf("failed to set up nftables egress table: %w", err)
	}
	f.active = true
	return nil
}

func (f *NFTables) setupScript(static []policy.Endpoint, redirect Redirect, withLog bool) string {
	mark := "0x" + strconv.FormatUint(uint64(f.Mark), 16)
	rules := []string{`oifname "lo" accept`}
	for _, p := range loopbackPrefixes {
		rules = append(rules, nftFamily(p.Addr().Is6())+" daddr "+p.String()+" accept")
	}
	rules = append(rules, "ct state established,related accept", nftNeighborDiscoveryRule())
	if f.Mark != 0 {
		rules = append(rules, "meta mark "+mark+" accept")
	}
	for _, e := range static {
		rules = append(rules, nftStaticRules(e)...)
	}
	if withLog {
		rules = append(rules, fmt.Sprintf(`log prefix "%s" group %d`, firewallNFLogPrefix, f.NFLogGroup))
	}
	if f.Mode == policy.ModeBlock {
		rules = append(rules, "meta l4proto tcp reject with tcp reset", "reject")
	}

	var b strings.Builder
	// create + delete removes leftovers of a previous run, the whole script is applied atomically
	b.WriteString("table " + nftTable + "\n")
	b.WriteString("delete table " + nftTable + "\n")
	b.WriteString("table " + nftTable + " {\n")
	b.WriteString("\tchain egress {\n")
	for _, rule := range rules {
		b.WriteString("\t\t" + rule + "\n")
	}
	b.WriteString("\t}\n")
	for _, hook := range []string{"output", "forward"} {
		b.WriteString(fmt.Sprintf("\tchain %s {\n\t\ttype filter hook %s priority 0; policy accept;\n\t\tjump egress\n\t}\n", hook, hook))
	}
	b.WriteString("\tchain nat_output {\n\t\ttype nat hook output priority -100; policy accept;\n")
	if f.Mark != 0 {
		b.WriteString("\t\tmeta mark " + mark + " return\n")
	}
	b.WriteString(fmt.Sprintf("\t\t%s daddr %s meta l4proto tcp redirect to :%d\n\t}\n", nftFamily(redirect.Prefix.Addr().Is6()), redirect.Prefix, redirect.Port))
	b.WriteString("}\n")
	return b.String()
}

func nftFamily(ipv6 bool) string {
	if ipv6 {
		return "ip6"
	}
	return "ip"
}

func nftNeighborDiscoveryRule() string {
	targets := make([]string, 0, len(neighborDiscoveryTargets))
	for _, p := range neighborDiscoveryTargets {
		targets = append(targets, p.String())
	}
	types := make([]string, 0, len(neighborDiscoveryTypes))
	for _, t := range neighborDiscoveryTypes {
		types = append(types, t.Name)
	}
	return "ip6 daddr { " + strings.Join(targets, ", ") + " } icmpv6 type { " + strings.Join(types, ", ") + " } accept"
}

func nftStaticRules(e policy.Endpoint) []string {
	match := nftFamily(e.Prefix.Addr().Is6()) + " daddr " + e.Prefix.String()
	if e.Port == 0 {
		return []string{match + " accept"}
	}
	port := strconv.Itoa(int(e.Port))
	return []string{
		match + " tcp dport " + port + " accept",
		match + " udp dport " + port + " accept",
	}
}

// Close removes the egress table
func (f *NFTables) Close() error {
	if !f.active {
		return nil
	}
	f.active = false
	return f.Run("delete table " + nftTable + "\n")
}
