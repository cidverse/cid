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
		"project-dir":       hc.projectDir,
		"work-dir":          filesystem.GetWorkingDirectory(),
		"user-id":           currentUser.Uid,
		"user-group-id":     currentUser.Gid,
		"user-login-name":   currentUser.Username,
		"user-display-name": currentUser.Name,
		"path-dist":         "",
		"path-temp":         "",
	}

	return c.JSON(http.StatusOK, res)
}
