package restapi

import (
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cidverseutils/ci"
	"github.com/labstack/echo/v4"
)

// configCurrent returns the configuration for the current action
func (hc *APIConfig) configCurrent(c echo.Context) error {
	host, _ := os.Hostname()
	currentUser, _ := user.Current()

	res := map[string]interface{}{
		// enable debugging
		"debug": false,
		// toggle debug output for specific parts of the process
		"log": map[string]string{
			"bin-helm": "debug",
		},
		// host
		"host_name":      host,
		"host_user_id":   currentUser.Uid,
		"host_user_name": currentUser.Username,
		"host_group_id":  currentUser.Gid,
		// paths
		"project_dir":  ci.ToUnixPath(hc.ProjectDir),
		"artifact_dir": ci.ToUnixPath(filepath.Join(hc.ProjectDir, ".dist")),
		"temp_dir":     constants.TempPathInContainer,
		// dynamic config
		"config": hc.ActionConfig,
	}

	return c.JSON(http.StatusOK, res)
}
