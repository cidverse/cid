package restapi

import (
	"net/http"
	"strconv"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/labstack/echo/v5"
)

// vcsCommits returns a list of commits between two refs
func (hc *APIConfig) vcsCommitsV1(c *echo.Context) error {
	limit := 0
	if c.QueryParam("limit") != "" {
		l, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return hc.handleError(c, http.StatusBadRequest, "parameter has a invalid value: limit", err.Error())
		}
		limit = l
	}

	result, err := hc.SDKClient.VCSCommitsV1(actionsdk.VCSCommitsRequest{
		FromHash:       c.QueryParam("from"),
		ToHash:         c.QueryParam("to"),
		IncludeChanges: c.QueryParam("changes") == "true",
		Limit:          limit,
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to query commits", err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// vcsCommitByHash retrieves information about a commit by hash
func (hc *APIConfig) vcsCommitByHashV1(c *echo.Context) error {
	result, err := hc.SDKClient.VCSCommitByHashV1(actionsdk.VCSCommitByHashRequest{
		Hash:           c.QueryParam("hash"),
		IncludeChanges: c.QueryParam("changes") == "true",
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to query commit", err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// vcsTags returns all tags
func (hc *APIConfig) vcsTagsV1(c *echo.Context) error {
	result, err := hc.SDKClient.VCSTagsV1()
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to query tags", err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// vcsTags returns all tags
func (hc *APIConfig) vcsReleasesV1(c *echo.Context) error {
	result, err := hc.SDKClient.VCSReleasesV1(actionsdk.VCSReleasesRequest{
		Type: c.QueryParam("type"),
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to query releases", err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// vcsDiff returns the diff between two refs
func (hc *APIConfig) vcsDiffV1(c *echo.Context) error {
	result, err := hc.SDKClient.VCSDiffV1(actionsdk.VCSDiffRequest{
		FromHash: c.QueryParam("from"),
		ToHash:   c.QueryParam("to"),
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to query diff", err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
