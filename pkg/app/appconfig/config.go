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
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Config struct {
	Version          string   `json:"version" env:"CLI_VERSION" validate:"required"`         // Version of this project
	JobTimeout       int      `json:"job_timeout" env:"JOB_TIMEOUT" validate:"required"`     // Timeout for the job in minutes
	JobRetries       int      `json:"job_retries" env:"JOB_RETRIES" validate:"required"`     // Number of retries for the job
	RunnerTags       []string `json:"runner_tags" env:"RUNNER_TAGS" validate:"required"`     // Tags for the runner jobs (e.g. "docker", "podman", ...)
	EgressPolicy     string   `json:"egress_policy" env:"EGRESS_POLICY" validate:"required"` // Egress policy for network traffic (block, audit, ...)
	ContainerRuntime string   `json:"container_runtime" env:"CONTAINER_RUNTIME" validate:"required,oneof=podman docker"`

	Workflows *orderedmap.OrderedMap[string, WorkflowConfig] `json:"workflows"`
}

type WorkflowConfig struct {
	Type                string   `json:"type"` // e.g. "cron", "dispatch", "manual"
	TriggerManual       bool     `json:"trigger_manual,omitempty"`
	TriggerSchedule     bool     `json:"trigger_cron,omitempty"`
	TriggerScheduleCron string   `json:"trigger_cron_schedule,omitempty"`
	TriggerPush         bool     `json:"trigger_push,omitempty"`
	TriggerPushBranches []string `json:"trigger_push_branches,omitempty"`
	TriggerPushTags     []string `json:"trigger_push_tags,omitempty"`
	TriggerPullRequest  bool     `json:"trigger_pull_request,omitempty"`
	EnvironmentPattern  string   `json:"environment_pattern,omitempty"` // EnvironmentPattern can be a regex or glob pattern to match which environments should be deployed from this workflow
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
	Version string `json:"version,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

func FormatDependencyReference(dep WorkflowDependency) string {
	if dep.Type == "oci-container" {
		if dep.Hash != "" {
			return fmt.Sprintf("%s@sha256:%s # %s", dep.Id, dep.Hash, dep.Version)
		} else {
			return fmt.Sprintf("%s:%s", dep.Id, dep.Version)
		}
	}

	return fmt.Sprintf("%s:%s", dep.Id, dep.Version)
}

func DefaultWorkflowConfig(defaultBranch string) *orderedmap.OrderedMap[string, WorkflowConfig] {
	workflowMap := orderedmap.New[string, WorkflowConfig]()
	workflowMap.Set("Main", WorkflowConfig{
		Type:                "main",
		TriggerManual:       true,
		TriggerPush:         true,
		TriggerPushBranches: []string{defaultBranch},
	})
	workflowMap.Set("Release", WorkflowConfig{
		Type:               "release",
		TriggerManual:      true,
		TriggerPush:        true,
		TriggerPushTags:    []string{"v[0-9]+.[0-9]+.[0-9]+"}, // try to use patterns that are compatible with regex (gitlab) and glob (github)
		EnvironmentPattern: "release-.*",
	})
	workflowMap.Set("Pull Request", WorkflowConfig{
		Type:               "pull-request",
		TriggerPullRequest: true,
		EnvironmentPattern: "pr-.*",
	})
	workflowMap.Set("Nightly", WorkflowConfig{
		Type:                "nightly",
		TriggerManual:       true,
		TriggerSchedule:     true,
		TriggerScheduleCron: "@weekly",
		EnvironmentPattern:  "nightly-.*",
	})

	return workflowMap
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
