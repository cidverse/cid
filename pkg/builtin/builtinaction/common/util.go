package common

import (
	"encoding/json"
	"fmt"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/go-playground/validator/v10"
)

// ParseAndValidateConfig parses JSON config, populates from env, and validates the struct.
func ParseAndValidateConfig(rawConfig string, env map[string]string, out any) error {
	// Parse JSON config if provided
	if rawConfig != "" {
		if err := json.Unmarshal([]byte(rawConfig), out); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Populate from env
	actionsdk.PopulateFromEnv(out, env)

	// Validate the struct
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(out); err != nil {
		return err
	}

	return nil
}

func MergeActionAccessNetwork(groups ...[]actionsdk.ActionAccessNetwork) []actionsdk.ActionAccessNetwork {
	var merged []actionsdk.ActionAccessNetwork
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return merged
}
