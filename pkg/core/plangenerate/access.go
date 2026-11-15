package plangenerate

import (
	"encoding/json"
	"fmt"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/egress/policy"
)

// StepAccess is the compact access definition persisted with each step of a plan.
// It keeps everything required to render and execute the step, but omits documentation-only fields of the action metadata.
type StepAccess struct {
	Environment []StepAccessEnv                    `json:"env,omitempty"`         // Environment variables that the step may access during execution
	Executables []actionsdk.ActionAccessExecutable `json:"executables,omitempty"` // Executables that the step may invoke during execution
	Network     StepAccessNetwork                  `json:"network,omitempty"`     // Network endpoints (short form, see policy.NetworkEntry) that the step may use during execution
	Resources   []actionsdk.ActionAccessResource   `json:"resources,omitempty"`   // Resources the step may access (e.g. releases, packages, ...)
}

type StepAccessEnv struct {
	Name     string `json:"name"`
	Pattern  bool   `json:"pattern,omitempty"`
	Required bool   `json:"required,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
}

// StepAccessNetwork is a list of network endpoints in the short form of policy.NetworkEntry: plain entries are just
// `host:port`, http rules / types / allowed ips are appended (e.g. `http://crl3.digicert.com:80|GET:/*.crl`)
type StepAccessNetwork []string

// UnmarshalJSON accepts both plain endpoint strings and the legacy {"host": "..."} object form
func (n *StepAccessNetwork) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make(StepAccessNetwork, 0, len(raw))
	for _, entry := range raw {
		var host string
		if err := json.Unmarshal(entry, &host); err == nil {
			result = append(result, host)
			continue
		}

		var legacy actionsdk.ActionAccessNetwork
		if err := json.Unmarshal(entry, &legacy); err != nil {
			return fmt.Errorf("invalid network access entry %s: %w", string(entry), err)
		}
		result = append(result, NetworkEntry(legacy))
	}

	*n = result
	return nil
}

// NewStepAccess converts the access definition of an action into the compact step access definition
func NewStepAccess(access actionsdk.ActionAccess) StepAccess {
	var env []StepAccessEnv
	for _, e := range access.Environment {
		env = append(env, StepAccessEnv{Name: e.Name, Pattern: e.Pattern, Required: e.Required, Secret: e.Secret})
	}

	return StepAccess{
		Environment: env,
		Executables: access.Executables,
		Network:     NetworkHosts(access.Network),
		Resources:   access.Resources,
	}
}

// NetworkHosts converts action network access entries into a list of endpoints (short form)
func NetworkHosts(network []actionsdk.ActionAccessNetwork) StepAccessNetwork {
	var result StepAccessNetwork
	for _, n := range network {
		result = append(result, NetworkEntry(n))
	}
	return result
}

// NetworkEntry converts an action network access entry into its short form, see policy.NetworkEntry
func NetworkEntry(n actionsdk.ActionAccessNetwork) string {
	entry := policy.NetworkEntry{Host: n.Host, Type: n.Type, AllowedIPs: n.AllowedIPs}
	for _, r := range n.HTTPRules {
		entry.HTTP.Rules = append(entry.HTTP.Rules, policy.HTTPRule{Methods: r.Methods, Paths: r.Paths})
	}
	if len(entry.HTTP.Rules) > 0 && entry.Type == policy.TypeDefault {
		entry.Type = policy.TypeHTTP
	}
	return entry.String()
}

// NetworkHostPorts reduces endpoints to their host part (host:port), for allow lists that only support hosts
// (harden-runner). Types, http rules and allowed ips are dropped, the result is deduplicated.
func NetworkHostPorts(network StepAccessNetwork) StepAccessNetwork {
	hosts := make(StepAccessNetwork, 0, len(network))
	for _, endpoint := range network {
		if entry, err := policy.ParseNetworkEntry(endpoint); err == nil {
			endpoint = entry.Host
		}
		hosts = append(hosts, endpoint)
	}
	return MergeNetworkHosts(hosts)
}

// MergeNetworkHosts concatenates endpoint lists into a new slice, deduplicating and preserving order
func MergeNetworkHosts(groups ...StepAccessNetwork) StepAccessNetwork {
	var result StepAccessNetwork
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, host := range group {
			if host == "" {
				continue
			}
			if _, exists := seen[host]; exists {
				continue
			}
			seen[host] = struct{}{}
			result = append(result, host)
		}
	}
	return result
}

// RemoveNetworkHosts returns the endpoints of network that are not contained in remove, preserving order
func RemoveNetworkHosts(network StepAccessNetwork, remove StepAccessNetwork) StepAccessNetwork {
	removeSet := make(map[string]struct{}, len(remove))
	for _, host := range remove {
		removeSet[host] = struct{}{}
	}

	var result StepAccessNetwork
	for _, host := range network {
		if _, exists := removeSet[host]; !exists {
			result = append(result, host)
		}
	}
	return result
}
