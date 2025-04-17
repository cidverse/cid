package appconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/elliotchance/orderedmap/v3"
)

type Config struct {
	Version      string `json:"version" env:"CLI_VERSION" validate:"required"`         // Version of this project
	JobTimeout   int    `json:"job_timeout" env:"JOB_TIMEOUT" validate:"required"`     // Timeout for the job in minutes
	EgressPolicy string `json:"egress_policy" env:"EGRESS_POLICY" validate:"required"` // Egress policy for network traffic (block, audit, ...)

	Workflows *orderedmap.OrderedMap[string, WorkflowConfig] `json:"workflows"`
}

type WorkflowConfig struct {
	Type                string   `json:"type"` // e.g. "cron", "dispatch", "manual"
	TriggerManual       bool     `json:"trigger_manual"`
	TriggerSchedule     bool     `json:"trigger_cron"`
	TriggerScheduleCron string   `json:"trigger_cron_schedule"`
	TriggerPush         bool     `json:"trigger_push"`
	TriggerPushBranches []string `json:"trigger_push_branches"`
	TriggerPushTags     []string `json:"trigger_push_tags"`
	TriggerPullRequest  bool     `json:"trigger_pull_request"`
	EnvironmentPattern  string   `json:"environment_pattern"` // EnvironmentPattern can be a regex or glob pattern to match which environments should be deployed from this workflow
}

func PreProcessWorkflowConfig(wfConfig WorkflowConfig, repo api.Repository) WorkflowConfig {
	if wfConfig.TriggerScheduleCron == "@daily" {
		wfConfig.TriggerScheduleCron = appcommon.GenerateCron("daily", strconv.FormatInt(repo.Id, 10))
	} else if wfConfig.TriggerScheduleCron == "@weekly" {
		wfConfig.TriggerScheduleCron = appcommon.GenerateCron("weekly", strconv.FormatInt(repo.Id, 10))
	} else if wfConfig.TriggerScheduleCron == "@monthly" {
		wfConfig.TriggerScheduleCron = appcommon.GenerateCron("monthly", strconv.FormatInt(repo.Id, 10))
	}

	return wfConfig
}

type WorkflowDependency struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Hash    string `json:"hash"`
	Version string `json:"version"`
}

// PersistPlan saves the workflow plan to a file in the specified project directory.
func PersistPlan(plan plangenerate.Plan, file string) error {
	err := os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for workflow plan [%s]: %w", file, err)
	}

	planBytes, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow plan [%s]: %w", file, err)
	}

	err = os.WriteFile(file, planBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write workflow plan [%s]: %w", file, err)
	}

	return nil
}

func LoadPlan(projectDirectory string, file string) (plangenerate.Plan, error) {
	planFile := filepath.Join(projectDirectory, file)
	planBytes, err := os.ReadFile(planFile)
	if err != nil {
		return plangenerate.Plan{}, fmt.Errorf("failed to read workflow plan [%s]: %w", file, err)
	}

	var plan plangenerate.Plan
	err = json.Unmarshal(planBytes, &plan)
	if err != nil {
		return plangenerate.Plan{}, fmt.Errorf("failed to unmarshal workflow plan [%s]: %w", file, err)
	}

	return plan, nil
}
