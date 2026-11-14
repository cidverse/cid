package appserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/cidverse/cid/pkg/app/apptask"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/labstack/echo/v5"
	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/singleflight"
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

// pipelineGenerator returns the full PipelineArtifact as JSON, or the raw content of one generated file if `file` is set.
func (s *Server) pipelineGenerator(c *echo.Context) error {
	projectId, err := strconv.ParseInt(c.QueryParam("project_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
	}
	file := c.QueryParam("file")

	responseData, err := s.resolvePipeline(c, projectId)
	if err != nil {
		return err
	}

	// return file content
	if file != "" {
		if content, ok := responseData.WorkflowContent[file]; ok {
			return c.String(http.StatusOK, content)
		}

		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("file %s not found in pipeline", file))
	}

	return c.JSON(http.StatusOK, responseData)
}

// resolvePipeline looks up the repository and returns its cached or freshly rendered pipeline; errors are echo HTTP errors.
func (s *Server) resolvePipeline(c *echo.Context, projectId int64) (*PipelineArtifact, error) {
	// lookup repo live, so new / changed / archived projects are picked up without a restart
	repo, err := s.lookupRepository(projectId)
	if err != nil {
		if errors.Is(err, api.ErrRepositoryNotFound) || errors.Is(err, errRepositoryNotAccessible) {
			slog.With("project_id", projectId).With("reason", err).Debug("Repository not found or not accessible")
			return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("repository with ID %d not found", projectId))
		}
		slog.With("project_id", projectId).With("err", err).Warn("Failed to look up repository")
		return nil, echo.NewHTTPError(http.StatusBadGateway, "failed to look up repository")
	}
	branch := repo.DefaultBranch
	key := cacheKey(constants.Version, repoInstance(repo, s.cfg.Platform), projectId, branch)

	// concurrent requests for the same key share one in-flight lookup/render; each request waits at most RenderTimeout.
	// A timed-out render keeps running and still populates the cache, so a retry is served from cache.
	ch := s.inflight.DoChan(key, func() (any, error) {
		return s.loadOrRenderPipeline(key, repo, branch)
	})
	timer := time.NewTimer(s.cfg.RenderTimeout)
	defer timer.Stop()

	var res singleflight.Result
	select {
	case res = <-ch:
	case <-timer.C:
		slog.With("repository", repo.Path).With("key", key).With("timeout", s.cfg.RenderTimeout).Warn("Timed out waiting for pipeline")
		return nil, echo.NewHTTPError(http.StatusGatewayTimeout, "pipeline generation timed out, retry later")
	case <-c.Request().Context().Done():
		return nil, c.Request().Context().Err()
	}
	if res.Err != nil {
		slog.With("repository", repo.Path).With("key", key).With("err", res.Err).Warn("Failed to provide pipeline")
		return nil, echo.NewHTTPError(http.StatusBadGateway, "failed to generate pipeline")
	}
	if res.Shared {
		slog.With("key", key).Debug("served pipeline from a concurrent in-flight request")
	}

	return res.Val.(*PipelineArtifact), nil
}

// loadOrRenderPipeline returns the cached pipeline for key, or renders and caches it.
// Runs detached from the request context, since the result is shared with other waiting requests.
// Only renders (cache misses) take a render slot.
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

	// render pipeline, limited to MaxConcurrentRenders in parallel
	s.renderSlots <- struct{}{}
	defer func() { <-s.renderSlots }()

	pipelineResult, err := s.processRepository(repo, branch)
	if err != nil {
		return nil, err
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

// processRepository renders the pipeline; panics become errors, since this runs outside echo's recover middleware.
func (s *Server) processRepository(repo api.Repository, branch string) (result apptask.WorkflowTaskResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while processing repository: %v", r)
		}
	}()

	platform := s.cfg.Platform
	slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("branch", branch).With("platform", platform.Name()).Info("rendering pipeline")
	result, err = appcore.ProcessRepository(platform, repo, true)
	if err != nil {
		return result, fmt.Errorf("failed to process repository: %w", err)
	}

	return result, nil
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

// repoInstance returns the id of the repository's platform instance (slugified hostname, e.g. gitlab-com), falling back to the platform slug.
func repoInstance(repo api.Repository, platform api.Platform) string {
	if repo.PlatformId != "" && repo.PlatformId != "unknown" {
		return repo.PlatformId
	}

	return platform.Slug()
}
