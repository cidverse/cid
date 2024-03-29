package restapi

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/cidverse/go-vcs"
	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/hashicorp/go-version"
	"github.com/labstack/echo/v4"
)

// vcsCommits returns a list of commits between two refs
func (hc *APIConfig) vcsCommits(c echo.Context) error {
	fromRef, err := vcsapi.NewVCSRefFromString(c.QueryParam("from"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "parameter has a invalid value: from",
			Details: err.Error(),
		})
	}

	toRef, err := vcsapi.NewVCSRefFromString(c.QueryParam("to"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "parameter has a invalid value: to",
			Details: err.Error(),
		})
	}

	includeChanges := c.QueryParam("changes")
	limit := 0
	if c.QueryParam("limit") != "" {
		var err error
		limit, err = strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, apiError{
				Status:  400,
				Title:   "invalid value for limit",
				Details: c.QueryParam("limit") + " is not a valid value!",
			})
		}
	}

	client, err := vcs.GetVCSClient(hc.ProjectDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to open vcs repository",
			Details: err.Error(),
		})
	}

	commits, err := client.FindCommitsBetween(fromRef, toRef, includeChanges == "true", limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to query commits",
			Details: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, commits)
}

// vcsCommitByHash retrieves information about a commit by hash
func (hc *APIConfig) vcsCommitByHash(c echo.Context) error {
	vcsCommitHash := c.Param("hash")
	includeChanges := c.QueryParam("changes")

	client, clientErr := vcs.GetVCSClient(hc.ProjectDir)
	if clientErr != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to open vcs repository",
			Details: clientErr.Error(),
		})
	}

	commit, err := client.FindCommitByHash(vcsCommitHash, includeChanges == "true")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to find commit by hash",
			Details: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, commit)
}

// vcsTags returns all tags
func (hc *APIConfig) vcsTags(c echo.Context) error {
	client, clientErr := vcs.GetVCSClient(hc.ProjectDir)
	if clientErr != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to open vcs repository",
			Details: clientErr.Error(),
		})
	}

	tags := client.GetTags()
	return c.JSON(http.StatusOK, tags)
}

// ByVersion implements sort.Interface
type ByVersion []*version.Version

func (a ByVersion) Len() int           { return len(a) }
func (a ByVersion) Less(i, j int) bool { return a[i].Compare(a[j]) > 0 }
func (a ByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// vcsTags returns all tags
func (hc *APIConfig) vcsReleases(c echo.Context) error {
	releaseType := c.QueryParam("type")

	client, clientErr := vcs.GetVCSClient(hc.ProjectDir)
	if clientErr != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to open vcs repository",
			Details: clientErr.Error(),
		})
	}

	var versions []*version.Version
	var versionToTag = make(map[string]vcsapi.VCSRef)
	for _, tag := range client.GetTags() {
		v, vErr := version.NewVersion(tag.Value)
		if vErr == nil {
			versions = append(versions, v)
			versionToTag[v.String()] = tag
		}
	}
	sort.Sort(ByVersion(versions))

	var releases []map[string]interface{}
	for _, v := range versions {
		release := map[string]interface{}{
			"version": v.String(),
			"ref":     versionToTag[v.String()],
		}

		if len(releaseType) > 0 {
			if releaseType == "stable" && len(v.Prerelease()) > 0 {
				continue
			} else if releaseType == "unstable" && v.Prerelease() == "" {
				continue
			} else if releaseType != "stable" && releaseType != "unstable" {
				return c.JSON(http.StatusBadRequest, apiError{
					Status:  400,
					Title:   "bad request",
					Details: "release type must be empty, stable or unstable",
				})
			}
		}

		releases = append(releases, release)
	}

	return c.JSON(http.StatusOK, releases)
}

// vcsDiff returns the diff between two refs
func (hc *APIConfig) vcsDiff(c echo.Context) error {
	fromRef, err := vcsapi.NewVCSRefFromString(c.QueryParam("from"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "parameter has a invalid value: from",
			Details: err.Error(),
		})
	}

	toRef, err := vcsapi.NewVCSRefFromString(c.QueryParam("to"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "parameter has a invalid value: from",
			Details: err.Error(),
		})
	}

	client, err := vcs.GetVCSClient(hc.ProjectDir)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to open vcs repository",
			Details: err.Error(),
		})
	}

	diff, err := client.Diff(fromRef, toRef)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, apiError{
			Status:  500,
			Title:   "failed to generate diff",
			Details: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, diff)
}
