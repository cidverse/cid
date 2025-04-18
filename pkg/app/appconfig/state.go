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
