package appconfig

import (
	"strconv"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

type Config struct {
	Version      string `json:"version" env:"CLI_VERSION" validate:"required"`         // Version of this project
	JobTimeout   int    `json:"job_timeout" env:"JOB_TIMEOUT" validate:"required"`     // Timeout for the job in minutes
	EgressPolicy string `json:"egress_policy" env:"EGRESS_POLICY" validate:"required"` // Egress policy for network traffic (block, audit, ...)

	Workflows map[string]WorkflowConfig `json:"workflows"`
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
