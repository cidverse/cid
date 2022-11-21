package restapi

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/user"
)

// projectInformation returns all available information about the current project
func (hc handlerConfig) projectInformation(c echo.Context) error {
	currentUser, _ := user.Current()

	res := map[string]interface{}{
		"project_dir":       hc.projectDir,
		"work_dir":          filesystem.GetWorkingDirectory(),
		"user_id":           currentUser.Uid,
		"user_group_id":     currentUser.Gid,
		"user_login_name":   currentUser.Username,
		"user_display_name": currentUser.Name,
		"path_dist":         "",
		"path_temp":         "",
	}

	return c.JSON(http.StatusOK, res)
}
