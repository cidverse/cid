package sonarqube

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

var apiClient = resty.New()

func init() {
	apiClient.SetDisableWarn(true)
}

type ProjectBranchesList struct {
	Branches []*Branch `json:"branches,omitempty"`
}

type Branch struct {
	AnalysisDate string  `json:"analysisDate,omitempty"`
	IsMain       bool    `json:"isMain,omitempty"`
	MergeBranch  string  `json:"mergeBranch,omitempty"`
	Name         string  `json:"name,omitempty"`
	Status       *Status `json:"status,omitempty"`
	Type         string  `json:"type,omitempty"`
}

type Status struct {
	Bugs              int64  `json:"bugs,omitempty"`
	CodeSmells        int64  `json:"codeSmells,omitempty"`
	QualityGateStatus string `json:"qualityGateStatus,omitempty"`
	Vulnerabilities   int64  `json:"vulnerabilities,omitempty"`
}

func createProject(server string, accessToken string, organization string, projectKey string, projectName string) error {
	resp, err := apiClient.R().
		SetQueryParams(map[string]string{
			"organization": organization,
			"project":      projectKey,
			"name":         projectName,
		}).
		SetHeader("Accept", "application/json").
		SetBasicAuth(accessToken, "").
		Post(server + "/api/projects/create")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("SonarQube deleteBranch failed - HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func getDefaultBranch(server string, accessToken string, projectKey string) (ProjectBranchesList, error) {
	resp, err := apiClient.R().
		SetQueryParams(map[string]string{
			"project": projectKey,
		}).
		SetHeader("Accept", "application/json").
		SetBasicAuth(accessToken, "").
		Get(server + "/api/project_branches/list")
	if err != nil {
		return ProjectBranchesList{}, err
	}
	if !resp.IsSuccess() {
		return ProjectBranchesList{}, fmt.Errorf("SonarQube deleteBranch failed - HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	result := ProjectBranchesList{}
	json.Unmarshal(resp.Body(), &result)

	return result, nil
}

func renameMainBranch(server string, accessToken string, projectKey string, name string) error {
	resp, err := apiClient.R().
		SetQueryParams(map[string]string{
			"project": projectKey,
			"name":    name,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(accessToken, "").
		Post(server + "/api/project_branches/rename")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("SonarQube deleteBranch failed - HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

func deleteBranch(server string, accessToken string, projectKey string, name string) error {
	resp, err := apiClient.R().
		SetQueryParams(map[string]string{
			"project": projectKey,
			"branch":  name,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(accessToken, "").
		Post(server + "/api/project_branches/delete")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("SonarQube deleteBranch failed - HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}
