package appserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/labstack/echo/v5"
)

// healthCheck returns a simple up status
func (s *Server) healthCheck(c *echo.Context) error {
	res := map[string]interface{}{
		"status": "up",
	}

	return c.JSON(http.StatusOK, res)
}

type PipelineArtifact struct {
	ProjectID       int64                    `json:"project_id"`
	RepoPath        string                   `json:"repo_path"`
	GeneratedAt     time.Time                `json:"generated_at"`
	WorkflowState   *appconfig.WorkflowState `json:"workflow_state"`
	WorkflowContent map[string]string        `json:"workflow_content"`
}

// pipelineGenerator
func (s *Server) pipelineGenerator(c *echo.Context) error {
	projectIdStr := c.QueryParam("project_id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
	}
	file := c.QueryParam("file")

	// storage
	bucketName := "cidverse-cid"
	objectName := fmt.Sprintf("%s/%d.json", strings.ToLower(s.cfg.Platform.Name()), projectId)
	var responseData *PipelineArtifact

	// check storage for cached artifact
	if s.cfg.StorageApi != nil {
		object, err := s.cfg.StorageApi.GetObject(context.Background(), bucketName, objectName)
		if err == nil {
			objectBytes, err := io.ReadAll(object)
			if err != nil {
				slog.Error("GetObject Read Error: " + err.Error())
			}
			err = json.Unmarshal(objectBytes, &responseData)
			if err != nil {
				slog.Error("GetObject Unmarshal Error: " + err.Error())
			}
		}
	} else {
		slog.Debug("Storage API not configured, skipping retrieval from cache")
	}

	if responseData == nil {
		// Lookup repo
		platform := s.cfg.Platform
		repo, ok := s.cfg.RepositoriesByID[projectId]
		if !ok {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("repository with ID %d not found", projectId))
		}

		// render pipeline
		slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("platform", platform.Name()).Info("running workflow update task")
		pipelineResult, err := appcore.ProcessRepository(platform, repo, true)
		if err != nil {
			slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).With("err", err).Warn("Failed to process repository")
		}

		// prepare response
		responseData = &PipelineArtifact{
			ProjectID:       projectId,
			RepoPath:        repo.Path,
			GeneratedAt:     time.Now(),
			WorkflowState:   pipelineResult.WorkflowState,
			WorkflowContent: pipelineResult.WorkflowContent,
		}
	}

	// return file content
	if file != "" {
		if content, ok := responseData.WorkflowContent[file]; ok {
			return c.String(200, content)
		} else {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("file %s not found in pipeline", file))
		}
	}

	// store result, if storage api is provided
	if s.cfg.StorageApi != nil {
		data, err := json.Marshal(responseData)
		if err != nil {
			slog.Error("JSON Marshal Error: " + err.Error())
		}

		err = s.cfg.StorageApi.PutObject(context.Background(), bucketName, objectName, bytes.NewReader(data), "application/json")
		if err != nil {
			slog.Error("PutObject Error: " + err.Error())
		}
	}

	return c.JSON(http.StatusOK, responseData)
}
