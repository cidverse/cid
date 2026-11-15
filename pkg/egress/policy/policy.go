package policy

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// Mode controls whether non-allowed egress is blocked or only reported
type Mode string

const (
	ModeBlock Mode = "block"
	ModeAudit Mode = "audit"
)

func ParseMode(s string) (Mode, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(s))) {
	case ModeBlock:
		return ModeBlock, nil
	case ModeAudit:
		return ModeAudit, nil
	default:
		return "", fmt.Errorf("invalid egress mode %q, allowed: block, audit", s)
	}
}

// Endpoint is a single allow list entry.
//
// Supported formats (port is optional, missing or `*` means any port):
//   - `github.com:443`
//   - `*.githubusercontent.com:443` (subdomains only, not the domain itself)
//   - `1.2.3.4:443`, `10.0.0.0/8`, `[2001:db8::/32]:443`
//
// Type, http rules and allowed ips are set via the short form / policy files, see NetworkEntry.
type Endpoint struct {
	Host     string       // normalized hostname (without wildcard prefix), empty for IP/CIDR entries
	Wildcard bool         // matches subdomains of Host
	Prefix   netip.Prefix // valid for IP/CIDR entries
	Port     uint16       // 0 = any port

	Type       string         // "" (tls sni / http host verified), "tcp" (raw relay) or "http" (request inspection), hostnames only
	HTTP       HTTPConfig     // type http only: request rules
	AllowedIPs []netip.Prefix // upstream addresses the host may resolve to, also allows private ranges
}

// endpoint types
const (
	TypeDefault = ""     // connection checks: tls sni / http host header must match the host
	TypeTCP     = "tcp"  // raw relay without protocol detection (databases, ssh, ...)
	TypeHTTP    = "http" // http request inspection (https is terminated with the inspection ca)
)

func (e Endpoint) String() string {
	host := e.Host
	if e.Wildcard {
		host = "*." + host
	}
	if e.Prefix.IsValid() {
		host = e.Prefix.String()
		if e.Prefix.IsSingleIP() {
			host = e.Prefix.Addr().String()
		}
	}
	port := "*"
	if e.Port != 0 {
		port = strconv.Itoa(int(e.Port))
	}
	return net.JoinHostPort(host, port)
}

// ParseEndpoint parses an allow list entry in the short form (see NetworkEntry), e.g. `github.com:443` or
// `http://api.github.com:443|GET:/repos/**`
func ParseEndpoint(s string) (Endpoint, error) {
	n, err := ParseNetworkEntry(s)
	if err != nil {
		return Endpoint{}, err
	}
	return n.Endpoint()
}

// parseHost parses the host part of an entry (hostname, wildcard, ip or cidr with an optional port)
func parseHost(s string) (Endpoint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Endpoint{}, fmt.Errorf("empty endpoint")
	}

	host, portStr := s, ""
	if h, p, err := net.SplitHostPort(s); err == nil {
		host, portStr = h, p
	}

	var e Endpoint
	if portStr != "" && portStr != "*" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil || port == 0 {
			return Endpoint{}, fmt.Errorf("invalid port in endpoint %q", s)
		}
		e.Port = uint16(port)
	}

	if strings.Contains(host, "/") {
		prefix, err := netip.ParsePrefix(host)
		if err != nil {
			return Endpoint{}, fmt.Errorf("invalid cidr in endpoint %q: %w", s, err)
		}
		e.Prefix = prefix.Masked()
		return e, nil
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		addr = addr.Unmap()
		e.Prefix = netip.PrefixFrom(addr, addr.BitLen())
		return e, nil
	}

	host = NormalizeName(host)
	if rest, ok := strings.CutPrefix(host, "*."); ok {
		e.Wildcard = true
		host = rest
	}
	if host == "" || strings.Contains(host, "*") {
		return Endpoint{}, fmt.Errorf("invalid hostname in endpoint %q", s)
	}
	e.Host = host
	return e, nil
}

// Policy is the parsed egress allow list
type Policy struct {
	Endpoints []Endpoint
}

func NewPolicy(entries []string) (*Policy, error) {
	endpoints, err := ParseEndpoints(entries)
	if err != nil {
		return nil, err
	}
	return NewPolicyFromEndpoints(endpoints)
}

// ParseEndpoints parses allow list entries, each entry may contain multiple whitespace separated endpoints
func ParseEndpoints(entries []string) ([]Endpoint, error) {
	var result []Endpoint
	for _, entry := range entries {
		for _, field := range strings.Fields(entry) {
			e, err := ParseEndpoint(field)
			if err != nil {
				return nil, err
			}
			result = append(result, e)
		}
	}
	return result, nil
}

// NewPolicyFromEndpoints validates the endpoints and creates a policy
func NewPolicyFromEndpoints(endpoints []Endpoint) (*Policy, error) {
	for _, e := range endpoints {
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("invalid endpoint %s: %w", e, err)
		}
	}
	return &Policy{Endpoints: endpoints}, nil
}

func (e Endpoint) Validate() error {
	switch e.Type {
	case TypeDefault, TypeTCP, TypeHTTP:
	default:
		return fmt.Errorf("invalid type %q, allowed: tcp, http", e.Type)
	}
	if e.Type != TypeDefault && e.Host == "" {
		return fmt.Errorf("type %s requires a hostname", e.Type)
	}
	if e.Type != TypeHTTP && !e.HTTP.isZero() {
		return fmt.Errorf("http settings require type http")
	}
	for _, r := range e.HTTP.Rules {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Match returns all hostname endpoints that allow the name on the port
func (p *Policy) Match(name string, port uint16) []Endpoint {
	var result []Endpoint
	for _, e := range p.MatchHost(name) {
		if e.Port == 0 || e.Port == port {
			result = append(result, e)
		}
	}
	return result
}

// MatchHost returns all hostname endpoints that match the name, on any port
func (p *Policy) MatchHost(name string) []Endpoint {
	name = NormalizeName(name)
	var result []Endpoint
	for _, e := range p.Endpoints {
		if e.matchesName(name) {
			result = append(result, e)
		}
	}
	return result
}

// HasInspection reports whether any endpoint requires tls interception (the inspection ca): a `type: http` endpoint on
// any port but 80. Plain http (port 80, e.g. crl downloads) is inspected without the ca.
func (p *Policy) HasInspection() bool {
	for _, e := range p.Endpoints {
		if e.Type == TypeHTTP && e.Port != 80 {
			return true
		}
	}
	return false
}

func (e Endpoint) matchesName(name string) bool {
	if e.Host == "" {
		return false
	}
	if e.Wildcard {
		return strings.HasSuffix(name, "."+e.Host)
	}
	return name == e.Host
}

// StaticEndpoints returns all IP/CIDR entries
func (p *Policy) StaticEndpoints() []Endpoint {
	var result []Endpoint
	for _, e := range p.Endpoints {
		if e.Prefix.IsValid() {
			result = append(result, e)
		}
	}
	return result
}

func NormalizeName(name string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".")
}
