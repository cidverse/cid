package appconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cidverse/cidverseutils/hash"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type WorkflowState struct {
	Workflows *orderedmap.OrderedMap[string, WorkflowData] `json:"workflows"`
}

type ChangeEntry struct {
	Workflow string
	Scope    string
	Message  string
}

func (w *WorkflowState) Hash() (string, error) {
	// marshal state
	stateBytes, err := json.Marshal(w.Workflows)

	// hash state
	h, err := hash.SHA256Hash(bytes.NewReader(stateBytes))
	if err != nil {
		return "", err
	}

	return h, nil
}

func (w *WorkflowState) CompareTo(other *WorkflowState) []ChangeEntry {
	var changes []ChangeEntry

	// w can be null, if this is the initial run for a project
	prevWorkflows := w.Workflows
	if prevWorkflows == nil {
		prevWorkflows = orderedmap.New[string, WorkflowData]()
	}

	// detect removed workflows
	for pair := prevWorkflows.Oldest(); pair != nil; pair = pair.Prev() {
		key := pair.Key
		if _, exists := other.Workflows.Get(key); !exists {
			changes = append(changes, ChangeEntry{Workflow: key, Scope: "workflow", Message: fmt.Sprintf("removed workflow [%s]", key)})
		}
	}

	// detect added workflows
	for pair := other.Workflows.Oldest(); pair != nil; pair = pair.Next() {
		key := pair.Key
		newWf := pair.Value
		oldWf, exists := prevWorkflows.Get(key)

		if !exists {
			changes = append(changes, ChangeEntry{Workflow: key, Scope: "workflow", Message: fmt.Sprintf("added workflow [%s]", key)})
			continue
		}

		// attributes
		addChange(&changes, key, "config", "JobTimeout", oldWf.JobTimeout, newWf.JobTimeout)
		addChange(&changes, key, "config", "DefaultBranch", oldWf.DefaultBranch, newWf.DefaultBranch)
		addChange(&changes, key, "config", "TriggerManual", oldWf.WorkflowConfig.TriggerManual, newWf.WorkflowConfig.TriggerManual)
		addChange(&changes, key, "config", "TriggerSchedule", oldWf.WorkflowConfig.TriggerSchedule, newWf.WorkflowConfig.TriggerSchedule)
		addChange(&changes, key, "config", "TriggerScheduleCron", oldWf.WorkflowConfig.TriggerScheduleCron, newWf.WorkflowConfig.TriggerScheduleCron)
		addChange(&changes, key, "config", "TriggerPush", oldWf.WorkflowConfig.TriggerPush, newWf.WorkflowConfig.TriggerPush)
		addSliceChange(&changes, key, "config", "TriggerPushBranches", oldWf.WorkflowConfig.TriggerPushBranches, newWf.WorkflowConfig.TriggerPushBranches)
		addSliceChange(&changes, key, "config", "TriggerPushTags", oldWf.WorkflowConfig.TriggerPushTags, newWf.WorkflowConfig.TriggerPushTags)
		addChange(&changes, key, "config", "TriggerPullRequest", oldWf.WorkflowConfig.TriggerPullRequest, newWf.WorkflowConfig.TriggerPullRequest)

		// Compare Dependencies
		for depKey, newDep := range newWf.WorkflowDependency {
			oldDep, exists := oldWf.WorkflowDependency[depKey]
			if !exists {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "dependency",
					Message:  fmt.Sprintf("added dependency [%s]", depKey),
				})
				continue
			}
			if oldDep.Version != newDep.Version {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "dependency",
					Message:  fmt.Sprintf("dependency [%s] version changed (%s → %s)", depKey, oldDep.Version, newDep.Version),
				})
			}
			if oldDep.Hash != newDep.Hash {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "dependency",
					Message:  fmt.Sprintf("dependency [%s] hash changed (%s → %s)", depKey, oldDep.Hash, newDep.Hash),
				})
			}
		}
		for depKey := range oldWf.WorkflowDependency {
			if _, exists := newWf.WorkflowDependency[depKey]; !exists {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "dependency",
					Message:  fmt.Sprintf("removed dependency [%s]", depKey),
				})
			}
		}

		// Compare Steps by Name
		oldSteps := make(map[string]struct{})
		for _, s := range oldWf.Plan.Steps {
			oldSteps[s.Name] = struct{}{}
		}
		newSteps := make(map[string]struct{})
		for _, s := range newWf.Plan.Steps {
			newSteps[s.Name] = struct{}{}
			if _, exists := oldSteps[s.Name]; !exists {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "step",
					Message:  fmt.Sprintf("added step: [%s]", s.Name),
				})
			}
		}
		for s := range oldSteps {
			if _, exists := newSteps[s]; !exists {
				changes = append(changes, ChangeEntry{
					Workflow: key,
					Scope:    "step",
					Message:  fmt.Sprintf("removed step: [%s]", s),
				})
			}
		}
	}

	return changes
}

func NewWorkflowState() WorkflowState {
	return WorkflowState{
		Workflows: orderedmap.New[string, WorkflowData](),
	}
}

func ReadWorkflowState(file string) (WorkflowState, error) {
	planBytes, err := os.ReadFile(file)
	if err != nil {
		return WorkflowState{}, fmt.Errorf("failed to read workflow plan [%s]: %w", file, err)
	}

	var state WorkflowState
	err = json.Unmarshal(planBytes, &state)
	if err != nil {
		return WorkflowState{}, fmt.Errorf("failed to unmarshal workflow plan [%s]: %w", file, err)
	}

	return state, nil
}

func WriteWorkflowState(state WorkflowState, file string) error {
	err := os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for workflow plan [%s]: %w", file, err)
	}

	stateBytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow plan [%s]: %w", file, err)
	}

	err = os.WriteFile(file, stateBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write workflow plan [%s]: %w", file, err)
	}

	return nil
}
