package policy

import (
	"fmt"
	"slices"
	"strings"
)

// NetworkEntry is an allow list entry in the shape of an action's Access.Network entry. It is used by policy files
// (yaml), action metadata and - as short form string - by the plan / state file and `--egress-allow`.
//
// Short form: `[<type>://]<host>[|<rule>]...[|ips=<prefix>,...]`
//
//	api.nuget.org:443                                       # tls sni / http host verified, relayed
//	tcp://db.example:5432                                   # raw relay
//	http://api.github.com:443                               # inspect requests, every request allowed
//	http://api.github.com:443|GET,HEAD:/repos/cidverse/**   # inspect requests, only allow matching ones
//	http://crl3.digicert.com:80|GET:/*.crl|HEAD
//	gitlab.internal.example:443|ips=10.20.0.0/16            # may resolve to these (private) addresses
//
// A rule is `<methods>:<paths>`, `<paths>` (any method, starts with `/`) or `<methods>` (any path), methods and paths
// are comma separated. Rules imply type http, the `http://` prefix is optional then (but always written by String).
// Plain entries (no type, rules or ips) are just the host, so existing state files keep their format.
type NetworkEntry struct {
	Host       string     `yaml:"host"`
	Type       string     `yaml:"type,omitempty"`
	HTTP       HTTPConfig `yaml:"http,omitempty"`
	AllowedIPs []string   `yaml:"allowed-ips,omitempty"`
}

const (
	shortFormTypeSeparator = "://"
	shortFormSeparator     = "|"
	shortFormIPsPrefix     = "ips="
)

// ParseNetworkEntry parses the short form of an allow list entry, see NetworkEntry
func ParseNetworkEntry(s string) (NetworkEntry, error) {
	s = strings.TrimSpace(s)
	var n NetworkEntry
	if typ, rest, ok := strings.Cut(s, shortFormTypeSeparator); ok {
		switch typ {
		case TypeTCP, TypeHTTP:
			n.Type = typ
		default:
			return NetworkEntry{}, fmt.Errorf("invalid type %q in endpoint %q, allowed: tcp, http", typ, s)
		}
		s = rest
	}

	parts := strings.Split(s, shortFormSeparator)
	n.Host = parts[0]
	for _, part := range parts[1:] {
		if ips, ok := strings.CutPrefix(part, shortFormIPsPrefix); ok {
			n.AllowedIPs = append(n.AllowedIPs, splitShortFormList(ips)...)
			continue
		}
		rule, err := parseShortFormRule(part)
		if err != nil {
			return NetworkEntry{}, fmt.Errorf("endpoint %q: %w", s, err)
		}
		n.HTTP.Rules = append(n.HTTP.Rules, rule)
	}
	if len(n.HTTP.Rules) > 0 && n.Type == TypeDefault {
		n.Type = TypeHTTP
	}
	return n, nil
}

func parseShortFormRule(s string) (HTTPRule, error) {
	if s == "" {
		return HTTPRule{}, fmt.Errorf("empty http rule")
	}
	methods, paths := "", s
	if i := strings.Index(s, ":"); i >= 0 && !strings.Contains(s[:i], "/") {
		methods, paths = s[:i], s[i+1:]
	} else if !strings.HasPrefix(s, "/") {
		methods, paths = s, ""
	}
	rule := HTTPRule{Methods: splitShortFormList(methods), Paths: splitShortFormList(paths)}
	return rule, rule.Validate()
}

func splitShortFormList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// String returns the short form of the entry, see NetworkEntry
func (n NetworkEntry) String() string {
	var b strings.Builder
	if n.Type != TypeDefault {
		b.WriteString(n.Type + shortFormTypeSeparator)
	}
	b.WriteString(n.Host)
	for _, r := range n.HTTP.Rules {
		b.WriteString(shortFormSeparator)
		methods, paths := strings.Join(r.Methods, ","), strings.Join(r.Paths, ",")
		switch {
		case methods != "" && paths != "":
			b.WriteString(methods + ":" + paths)
		case paths != "":
			b.WriteString(paths)
		case methods != "":
			b.WriteString(methods)
		default:
			b.WriteString("*") // a rule without methods and paths allows every request
		}
	}
	if len(n.AllowedIPs) > 0 {
		b.WriteString(shortFormSeparator + shortFormIPsPrefix + strings.Join(n.AllowedIPs, ","))
	}
	return b.String()
}

// Endpoint parses and validates the entry
func (n NetworkEntry) Endpoint() (Endpoint, error) {
	e, err := parseHost(n.Host)
	if err != nil {
		return Endpoint{}, err
	}
	e.Type = n.Type
	e.HTTP = n.HTTP
	for _, ip := range n.AllowedIPs {
		prefix, err := parsePrefixOrAddr(ip)
		if err != nil {
			return Endpoint{}, fmt.Errorf("invalid allowed-ips entry %q: %w", ip, err)
		}
		e.AllowedIPs = append(e.AllowedIPs, prefix)
	}
	if err = e.Validate(); err != nil {
		return Endpoint{}, fmt.Errorf("%s: %w", n.Host, err)
	}
	return e, nil
}

// Rule returns the endpoint in the short form, it names the policy entry that allowed a request in reports
func (e Endpoint) Rule() string {
	n := NetworkEntry{Host: e.String(), Type: e.Type, HTTP: e.HTTP}
	for _, prefix := range e.AllowedIPs {
		n.AllowedIPs = append(n.AllowedIPs, prefix.String())
	}
	return n.String()
}

// Rules returns the distinct rules of the endpoints, joined with ", "
func Rules(endpoints []Endpoint) string {
	var rules []string
	for _, e := range endpoints {
		if rule := e.Rule(); !slices.Contains(rules, rule) {
			rules = append(rules, rule)
		}
	}
	return strings.Join(rules, ", ")
}
