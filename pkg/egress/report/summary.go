package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// FileVersion is the version of the summary file format
const FileVersion = 1

// MaxRequestsPerRule caps the distinct requests listed per rule (chatty apis, e.g. one path per npm package), the rest
// is counted in RuleSummary.Other
const MaxRequestsPerRule = 100

// File is the egress summary of a filter run (`egress.json`): allowed traffic grouped by the policy entry that allowed
// it, traffic outside the policy and policy entries without traffic.
type File struct {
	Version    int           `json:"version"`
	Mode       string        `json:"mode"`            // block or audit
	Inspection bool          `json:"inspection"`      // tls of type http endpoints was terminated with the inspection ca
	Steps      []string      `json:"steps,omitempty"` // steps that ran under the filter, the summary covers all of them
	Started    time.Time     `json:"started"`
	Finished   time.Time     `json:"finished"`
	Rules      []RuleSummary `json:"rules"`            // allowed traffic per policy entry
	Denied     []Event       `json:"denied,omitempty"` // traffic outside the policy (blocked, or passed in audit mode)
	Unused     []string      `json:"unused,omitempty"` // hostname policy entries without traffic (static ip entries are not observed)
}

// RuleSummary is the allowed traffic of one policy entry (or several entries that matched the same requests)
type RuleSummary struct {
	Rule     string       `json:"rule"`
	Count    int          `json:"count"`           // all requests / connections
	Requests []Event      `json:"requests"`        // distinct requests / connections, most frequent first
	Other    *OtherEvents `json:"other,omitempty"` // requests beyond MaxRequestsPerRule
}

// OtherEvents counts the requests that are not listed individually
type OtherEvents struct {
	Distinct int `json:"distinct"`
	Count    int `json:"count"`
}

// NewFile builds the summary from the reporter's events. rules are the hostname policy entries (short form), entries
// that allowed nothing are listed as unused.
func NewFile(events []Event, rules []string) File {
	byRule := map[string][]Event{}
	file := File{Version: FileVersion, Rules: []RuleSummary{}}
	for _, e := range events {
		if e.Action != EventActionAllowed {
			file.Denied = append(file.Denied, e)
			continue
		}
		byRule[e.Rule] = append(byRule[e.Rule], e)
	}

	used := map[string]bool{}
	for rule, requests := range byRule {
		used[rule] = true
		summary := RuleSummary{Rule: rule}
		sort.SliceStable(requests, func(i, j int) bool {
			if requests[i].Count != requests[j].Count {
				return requests[i].Count > requests[j].Count
			}
			return requests[i].key() < requests[j].key()
		})
		for i, e := range requests {
			summary.Count += e.Count
			if i >= MaxRequestsPerRule {
				if summary.Other == nil {
					summary.Other = &OtherEvents{}
				}
				summary.Other.Distinct++
				summary.Other.Count += e.Count
				continue
			}
			e.Rule = "" // the group names the rule
			summary.Requests = append(summary.Requests, e)
		}
		file.Rules = append(file.Rules, summary)
	}
	sort.Slice(file.Rules, func(i, j int) bool { return file.Rules[i].Rule < file.Rules[j].Rule })

	for _, rule := range rules {
		if !used[rule] && !usedInCombination(rule, used) {
			file.Unused = append(file.Unused, rule)
		}
	}
	return file
}

// usedInCombination reports whether the rule allowed traffic together with other entries ("a, b", see policy.Rules;
// short form rules never contain ", ")
func usedInCombination(rule string, used map[string]bool) bool {
	for combined := range used {
		if slices.Contains(strings.Split(combined, ", "), rule) {
			return true
		}
	}
	return false
}

// Write writes the summary as indented json, creating the parent directory
func (f File) Write(path string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err = os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write egress summary %s: %w", path, err)
	}
	return nil
}
