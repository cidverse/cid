package firewall

import (
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/rs/zerolog/log"
)

const (
	firewallChain       = "CID-EGRESS"
	firewallNFLogPrefix = "cid-egress"
)

// firewallHooks are the builtin chains that jump into the egress chain.
// OUTPUT covers the job itself, FORWARD covers nested containers on a bridge network (podman / netavark).
var firewallHooks = []string{"OUTPUT", "FORWARD"}

// loopbackPrefixes are always accepted by destination, not only by `oifname lo`: after the nat output hook redirects a
// connection to the proxy, the later filter hook still sees the original output interface (e.g. eth0) for the first
// packet, only the destination is rewritten to 127.0.0.1 / ::1.
var loopbackPrefixes = []netip.Prefix{
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("::1/128"),
}

// neighborDiscoveryTargets are the link-local scopes ipv6 neighbor discovery / multicast listener reports are sent to.
var neighborDiscoveryTargets = []netip.Prefix{
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff02::/16"),
}

// neighborDiscoveryTypes are the kernel generated icmpv6 messages ipv6 needs to reach the gateway (RFC 4861, RFC 3810),
// accepted without a policy entry or report. Rejecting them breaks ipv6, incl. the filter's own upstream connections
// (the kernel does not mark these packets).
var neighborDiscoveryTypes = []struct {
	Name string // nftables keyword
	Type int    // icmpv6 type, for ip6tables
}{
	{"nd-router-solicit", 133},
	{"nd-neighbor-solicit", 135},
	{"nd-neighbor-advert", 136},
	{"mld-listener-report", 131},
	{"mld2-listener-report", 143},
}

// Firewall restricts egress to the transparent proxy and static ip / cidr endpoints
type Firewall interface {
	// Setup installs the deny-by-default rules, allows the static ip / cidr endpoints and redirects tcp connections to
	// the synthetic prefix to the local transparent proxy
	Setup(static []policy.Endpoint, redirect Redirect) error
	// Close removes all rules
	Close() error
	// Snapshot returns the current firewall rules (without counters) as text, used to detect modifications after Setup.
	// It covers more than the filter's own rules: rules elsewhere can bypass them (e.g. set the filter's socket mark or
	// accept before the filter's jump in iptables).
	Snapshot() (string, error)
}

// OutputRunner executes a read-only firewall command and returns its output, replaceable for tests
type OutputRunner func(name string, args ...string) (string, error)

// ExecOutput executes the read-only firewall command on the host
func ExecOutput(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

// Redirect sends tcp connections to a prefix to a local port
type Redirect struct {
	Prefix netip.Prefix
	Port   uint16
}

const (
	BackendAuto     = "auto"
	BackendNFTables = "nftables"
	BackendIPTables = "iptables"
)

// New creates the firewall for a backend (auto = nftables if usable, otherwise iptables) that modifies the host
func New(backend string, mode policy.Mode, mark uint32, nflogGroup uint16) (Firewall, error) {
	if backend == "" || backend == BackendAuto {
		detected, err := detectFirewallBackend()
		if err != nil {
			return nil, err
		}
		backend = detected
	}
	log.Info().Str("backend", backend).Msg("using firewall backend")

	switch backend {
	case BackendNFTables:
		return &NFTables{Mode: mode, Mark: mark, NFLogGroup: nflogGroup, Run: ExecNFT, List: ExecOutput}, nil
	case BackendIPTables:
		return &IPTables{Mode: mode, Mark: mark, NFLogGroup: nflogGroup, Run: ExecRunner, List: ExecOutput}, nil
	default:
		return nil, fmt.Errorf("invalid firewall backend %q, allowed: auto, nftables, iptables", backend)
	}
}

func detectFirewallBackend() (string, error) {
	if _, err := exec.LookPath("nft"); err == nil {
		// verifies kernel support and permissions, without changing anything
		out, err := exec.Command("nft", "list", "tables").CombinedOutput()
		if err == nil {
			return BackendNFTables, nil
		}
		log.Debug().Err(err).Str("output", strings.TrimSpace(string(out))).Msg("nft is installed but not usable")
	}
	if _, err := exec.LookPath("iptables"); err == nil {
		return BackendIPTables, nil
	}
	return "", errors.New("no usable firewall backend found, install nftables (nft) or iptables")
}

var ErrUnsupportedPlatform = errors.New("not supported on this platform")
