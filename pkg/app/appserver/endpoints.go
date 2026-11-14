package appserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/labstack/echo/v5"
	"github.com/minio/minio-go/v7"
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
	Branch          string                   `json:"branch"`
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

	// lookup repo
	repo, ok := s.cfg.RepositoriesByID[projectId]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("repository with ID %d not found", projectId))
	}
	branch := repo.DefaultBranch
	key := cacheKey(constants.Version, repoInstance(repo, s.cfg.Platform), projectId, branch)

	// concurrent requests for the same key block until the first one is done and share its result
	result, err, shared := s.inflight.Do(key, func() (any, error) {
		return s.loadOrRenderPipeline(key, repo, branch)
	})
	if err != nil {
		slog.With("repository", repo.Path).With("key", key).With("err", err).Warn("Failed to provide pipeline")
		return echo.NewHTTPError(http.StatusBadGateway, "failed to generate pipeline")
	}
	if shared {
		slog.With("key", key).Debug("served pipeline from a concurrent in-flight request")
	}
	responseData := result.(*PipelineArtifact)

	// return file content
	if file != "" {
		if content, ok := responseData.WorkflowContent[file]; ok {
			return c.String(http.StatusOK, content)
		}

		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("file %s not found in pipeline", file))
	}

	return c.JSON(http.StatusOK, responseData)
}

// loadOrRenderPipeline returns the cached pipeline for key, or renders and caches it.
// Runs detached from the request context, since the result is shared with other waiting requests.
func (s *Server) loadOrRenderPipeline(key string, repo api.Repository, branch string) (*PipelineArtifact, error) {
	ctx := context.Background()

	// check storage for cached artifact
	if s.cfg.StorageApi != nil {
		cached, err := s.readCachedPipeline(ctx, key)
		if err != nil {
			slog.With("key", key).With("err", err).Warn("Failed to read cached pipeline, rendering instead")
		} else if cached != nil {
			return cached, nil
		}
	} else {
		slog.Debug("Storage API not configured, skipping retrieval from cache")
	}

	// render pipeline
	platform := s.cfg.Platform
	slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("branch", branch).With("platform", platform.Name()).Info("rendering pipeline")
	pipelineResult, err := appcore.ProcessRepository(platform, repo, true)
	if err != nil {
		return nil, fmt.Errorf("failed to process repository: %w", err)
	}

	artifact := &PipelineArtifact{
		ProjectID:       repo.Id,
		RepoPath:        repo.Path,
		Branch:          branch,
		GeneratedAt:     time.Now(),
		WorkflowState:   pipelineResult.WorkflowState,
		WorkflowContent: pipelineResult.WorkflowContent,
	}

	// store result, if storage api is provided
	if s.cfg.StorageApi != nil {
		data, err := json.Marshal(artifact)
		if err != nil {
			slog.With("key", key).With("err", err).Error("Failed to marshal pipeline for cache")
		} else if err = s.cfg.StorageApi.PutObject(ctx, s.cfg.StorageBucket, key, bytes.NewReader(data), "application/json"); err != nil {
			slog.With("key", key).With("err", err).Error("Failed to store pipeline in cache")
		}
	}

	return artifact, nil
}

// readCachedPipeline returns the cached pipeline, nil if the key does not exist, or an error if the cache could not be read.
func (s *Server) readCachedPipeline(ctx context.Context, key string) (*PipelineArtifact, error) {
	object, err := s.cfg.StorageApi.GetObject(ctx, s.cfg.StorageBucket, key)
	if err != nil {
		return nil, err
	}
	defer object.Close()

	// minio reports a missing key on the first read, not on GetObject
	objectBytes, err := io.ReadAll(object)
	if err != nil {
		if minio.ToErrorResponse(err).Code == minio.NoSuchKey {
			return nil, nil
		}
		return nil, err
	}

	var artifact PipelineArtifact
	if err = json.Unmarshal(objectBytes, &artifact); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached pipeline: %w", err)
	}

	return &artifact, nil
}

// cacheKey builds the storage key <cid-version>/<instance>/<project>/<branch>; each segment is path-escaped so branches like "feature/x" stay one segment.
func cacheKey(version string, instance string, projectId int64, branch string) string {
	return strings.Join([]string{
		url.PathEscape(version),
		url.PathEscape(instance),
		strconv.FormatInt(projectId, 10),
		url.PathEscape(branch),
	}, "/")
}

// repoInstance returns the host of the repository's platform instance (e.g. gitlab.com), falling back to the platform slug.
func repoInstance(repo api.Repository, platform api.Platform) string {
	if u, err := url.Parse(repo.URL); err == nil && u.Host != "" {
		return strings.ToLower(u.Host)
	}

	return platform.Slug()
}
