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

// CommandRunner executes a firewall command, replaceable for tests
type CommandRunner func(name string, args ...string) error

// ExecRunner executes the firewall command on the host
func ExecRunner(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// IPTables manages dedicated iptables / ip6tables chains that only allow egress through the transparent proxy, the
// fallback when nftables is not available.
//
// filter chain (jumped to from OUTPUT and FORWARD):
//
//	-o lo -j RETURN
//	-d 127.0.0.0/8 -j RETURN / -d ::1 -j RETURN          (connections redirected to the proxy, see loopbackPrefixes)
//	-m conntrack --ctstate ESTABLISHED,RELATED -j RETURN
//	-d fe80::/10,ff02::/16 -p ipv6-icmp --icmpv6-type <nd/mld> -j RETURN  (ip6tables only, ipv6 neighbor discovery)
//	-m mark --mark <mark> -j RETURN                     (the egress filter's own upstream connections)
//	-d <static-cidr> ... -j RETURN                      (static ip / cidr entries)
//	-j NFLOG --nflog-group <group>                      (report everything else)
//	-p tcp -j REJECT --reject-with tcp-reset            (block mode only)
//	-j REJECT                                           (block mode only)
//
// nat chain (jumped to from OUTPUT): marked connections return, tcp to the synthetic prefix is redirected to the proxy.
type IPTables struct {
	Mode       policy.Mode
	Mark       uint32
	NFLogGroup uint16        // 0 = disabled
	Run        CommandRunner // required, use ExecRunner to modify the host firewall (no default, so tests can't touch the host by accident)
	List       OutputRunner  // required for Snapshot, use ExecOutput

	active []string // binaries whose chains were set up
}

// Setup creates the egress chains for ipv4 and ipv6. IPv4 is required, IPv6 is skipped with a warning if unavailable.
func (f *IPTables) Setup(static []policy.Endpoint, redirect Redirect) error {
	if f.Run == nil {
		return errors.New("iptables: no command runner configured")
	}

	for _, family := range []struct {
		bin  string
		ipv6 bool
	}{{"iptables", false}, {"ip6tables", true}} {
		if err := f.setupFamily(family.bin, family.ipv6, static, redirect); err != nil {
			f.teardownFamily(family.bin)
			if !family.ipv6 {
				return fmt.Errorf("failed to set up %s egress chain: %w", family.bin, err)
			}
			log.Warn().Err(err).Msg("failed to set up ip6tables egress chain, ipv6 egress is NOT filtered")
			continue
		}
		f.active = append(f.active, family.bin)
	}
	return nil
}

func (f *IPTables) setupFamily(bin string, ipv6 bool, static []policy.Endpoint, redirect Redirect) error {
	f.teardownFamily(bin) // remove leftovers of a previous run
	mark := "0x" + strconv.FormatUint(uint64(f.Mark), 16)

	commands := [][]string{{"-N", firewallChain}}
	add := func(rule ...string) { commands = append(commands, append([]string{"-A", firewallChain}, rule...)) }
	add("-o", "lo", "-j", "RETURN")
	for _, p := range loopbackPrefixes {
		if p.Addr().Is6() == ipv6 {
			add("-d", p.String(), "-j", "RETURN")
		}
	}
	add("-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "RETURN")
	if ipv6 {
		for _, target := range neighborDiscoveryTargets {
			for _, t := range neighborDiscoveryTypes {
				add("-d", target.String(), "-p", "ipv6-icmp", "--icmpv6-type", strconv.Itoa(t.Type), "-j", "RETURN")
			}
		}
	}
	if f.Mark != 0 {
		add("-m", "mark", "--mark", mark, "-j", "RETURN")
	}
	for _, e := range static {
		if e.Prefix.Addr().Is6() != ipv6 {
			continue
		}
		if e.Port == 0 {
			add("-d", e.Prefix.String(), "-j", "RETURN")
			continue
		}
		port := strconv.Itoa(int(e.Port))
		add("-d", e.Prefix.String(), "-p", "tcp", "--dport", port, "-j", "RETURN")
		add("-d", e.Prefix.String(), "-p", "udp", "--dport", port, "-j", "RETURN")
	}
	if err := f.runAll(bin, commands); err != nil {
		return err
	}

	// optional, kernels / images without the NFLOG target still filter
	if f.NFLogGroup != 0 {
		if err := f.Run(bin, "-w", "-A", firewallChain, "-j", "NFLOG", "--nflog-group", strconv.Itoa(int(f.NFLogGroup)), "--nflog-prefix", firewallNFLogPrefix); err != nil {
			log.Warn().Err(err).Str("binary", bin).Msg("failed to add NFLOG rule, blocked connections will not be reported")
		}
	}

	commands = nil
	if f.Mode == policy.ModeBlock {
		add("-p", "tcp", "-j", "REJECT", "--reject-with", "tcp-reset")
		add("-j", "REJECT")
	}
	for _, hook := range firewallHooks {
		commands = append(commands, []string{"-I", hook, "1", "-j", firewallChain})
	}
	if redirect.Prefix.Addr().Is6() == ipv6 {
		commands = append(commands, []string{"-t", "nat", "-N", firewallChain})
		if f.Mark != 0 {
			commands = append(commands, []string{"-t", "nat", "-A", firewallChain, "-m", "mark", "--mark", mark, "-j", "RETURN"})
		}
		commands = append(commands,
			[]string{"-t", "nat", "-A", firewallChain, "-d", redirect.Prefix.String(), "-p", "tcp", "-j", "REDIRECT", "--to-ports", strconv.Itoa(int(redirect.Port))},
			[]string{"-t", "nat", "-I", "OUTPUT", "1", "-j", firewallChain},
		)
	}
	return f.runAll(bin, commands)
}

func (f *IPTables) runAll(bin string, commands [][]string) error {
	for _, args := range commands {
		if err := f.Run(bin, append([]string{"-w"}, args...)...); err != nil {
			return err
		}
	}
	return nil
}

// Close removes the egress chains and all jumps into them
func (f *IPTables) Close() error {
	for _, bin := range f.active {
		f.teardownFamily(bin)
	}
	f.active = nil
	return nil
}

func (f *IPTables) teardownFamily(bin string) {
	for _, table := range []struct {
		name  string
		hooks []string
	}{{"filter", firewallHooks}, {"nat", []string{"OUTPUT"}}} {
		for _, hook := range table.hooks {
			// remove all jumps, a previous run might have left more than one
			for i := 0; i < 16; i++ {
				if err := f.Run(bin, "-w", "-t", table.name, "-D", hook, "-j", firewallChain); err != nil {
					break
				}
			}
		}
		_ = f.Run(bin, "-w", "-t", table.name, "-F", firewallChain)
		_ = f.Run(bin, "-w", "-t", table.name, "-X", firewallChain)
	}
}

// snapshotTables are listed by Snapshot: rules in any of them can bypass the egress chain (accept before the jump, set
// the socket mark, notrack)
var snapshotTables = []string{"filter", "nat", "mangle", "raw"}

// Snapshot returns the rules (`-S`, without counters) of all tables for every family that was set up. A table that
// can't be listed (module not available) is recorded as unavailable instead of failing.
func (f *IPTables) Snapshot() (string, error) {
	if f.List == nil {
		return "", errors.New("iptables: no output runner configured")
	}
	if len(f.active) == 0 {
		return "", errors.New("iptables: not set up")
	}
	var b strings.Builder
	for _, bin := range f.active {
		for _, table := range snapshotTables {
			b.WriteString("# " + bin + " " + table + "\n")
			out, err := f.List(bin, "-w", "-t", table, "-S")
			if err != nil {
				b.WriteString("# unavailable\n")
				continue
			}
			b.WriteString(out)
		}
	}
	return b.String(), nil
}
