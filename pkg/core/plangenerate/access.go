package plangenerate

import (
	"encoding/json"
	"fmt"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

// StepAccess is the compact access definition persisted with each step of a plan.
// It keeps everything required to render and execute the step, but omits documentation-only fields of the action metadata.
type StepAccess struct {
	Environment []StepAccessEnv                    `json:"env,omitempty"`         // Environment variables that the step may access during execution
	Executables []actionsdk.ActionAccessExecutable `json:"executables,omitempty"` // Executables that the step may invoke during execution
	Network     StepAccessNetwork                  `json:"network,omitempty"`     // Network endpoints (host:port) that the step may use during execution
	Resources   []actionsdk.ActionAccessResource   `json:"resources,omitempty"`   // Resources the step may access (e.g. releases, packages, ...)
}

type StepAccessEnv struct {
	Name     string `json:"name"`
	Pattern  bool   `json:"pattern,omitempty"`
	Required bool   `json:"required,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
}

// StepAccessNetwork is a list of network endpoints (host:port)
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
		result = append(result, legacy.Host)
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

// NetworkHosts converts action network access entries into a list of endpoints
func NetworkHosts(network []actionsdk.ActionAccessNetwork) StepAccessNetwork {
	var result StepAccessNetwork
	for _, n := range network {
		result = append(result, n.Host)
	}
	return result
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
