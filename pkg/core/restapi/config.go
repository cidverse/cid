package restapi

import (
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/labstack/echo/v4"
)

// configCurrent returns the configuration for the current action
func (hc *handlerConfig) configCurrent(c echo.Context) error {
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
		"project_dir":  cihelper.ToUnixPath(hc.projectDir),
		"artifact_dir": cihelper.ToUnixPath(filepath.Join(hc.projectDir, ".dist")),
		"temp_dir":     cihelper.ToUnixPath(filepath.Join(hc.projectDir, ".tmp")),
		// dynamic config
		"config": hc.actionConfig,
	}

	return c.JSON(http.StatusOK, res)
}
