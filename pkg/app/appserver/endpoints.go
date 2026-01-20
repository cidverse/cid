package appserver

import (
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"

	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/labstack/echo/v5"
)

// healthCheck returns a simple up status
func (s *Server) healthCheck(c echo.Context) error {
	res := map[string]interface{}{
		"status": "up",
	}

	return c.JSON(http.StatusOK, res)
}

// pipelineGenerator
func (s *Server) pipelineGenerator(c echo.Context) error {
	projectIdStr := c.QueryParam("project_id")
	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
	}
	file := c.QueryParam("file")

	// TODO: retrieve result from storage if available (e.g. s3)
	// var storageApi storageapi.API

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

	// return file content
	if file != "" {
		if content, ok := pipelineResult.WorkflowContent[file]; ok {
			return c.String(200, content)
		} else {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("file %s not found in pipeline", file))
		}
	}

	// return data
	res := map[string]interface{}{
		"status":  repo.Id,
		"path":    repo.Path,
		"content": pipelineResult.WorkflowContent,
		"state":   pipelineResult.WorkflowState,
		"files":   slices.Collect(maps.Keys(pipelineResult.WorkflowContent)),
	}
	return c.JSON(http.StatusOK, res)
}
