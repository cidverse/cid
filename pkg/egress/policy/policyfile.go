package policy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"

	"go.yaml.in/yaml/v4"
)

// policyFile is the yaml policy format. Each network entry has the shape of an action's Access.Network entry (`host`),
// extended with the connection type and http request rules:
//
//	network:
//	  - host: proxy.golang.org:443           # same as action metadata: tls sni verified, relayed
//	  - host: "*.githubusercontent.com:443"  # subdomains
//	  - host: api.github.com:443
//	    type: http                           # inspect requests (tls terminated with the inspection ca)
//	    http:
//	      rules:                             # allow rules only, everything else is denied
//	        - methods: [GET, HEAD]
//	          paths: ["/repos/cidverse/**"]
//	        - methods: [POST]
//	          paths: ["/repos/cidverse/*/issues/*/comments"]
//	  - host: db.example:5432
//	    type: tcp                            # raw relay, no protocol detection
//	  - host: gitlab.internal.example:443
//	    allowed-ips: [10.20.0.0/16]          # may resolve to these (private) addresses
//	  - host: 10.0.0.0/8                     # static network
type policyFile struct {
	Network []NetworkEntry `yaml:"network"`
}

// ReadPolicyFile reads endpoints from a yaml policy file
func ReadPolicyFile(path string) ([]Endpoint, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParsePolicy(content)
}

// ParsePolicy parses endpoints from a yaml policy document, unknown fields are rejected
func ParsePolicy(content []byte) ([]Endpoint, error) {
	var file policyFile
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("invalid policy: %w", err)
	}

	result := make([]Endpoint, 0, len(file.Network))
	for i, entry := range file.Network {
		e, err := entry.Endpoint()
		if err != nil {
			return nil, fmt.Errorf("network entry %d: %w", i, err)
		}
		result = append(result, e)
	}
	return result, nil
}

func parsePrefixOrAddr(s string) (netip.Prefix, error) {
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return prefix.Masked(), nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Prefix{}, err
	}
	addr = addr.Unmap()
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}
