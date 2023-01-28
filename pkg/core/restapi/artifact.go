package restapi

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/cidverse/cid/pkg/core/state"
	"github.com/labstack/echo/v4"
)

// artifactList lists all generated reports
func (hc *handlerConfig) artifactList(c echo.Context) error {
	var result = make([]state.ActionArtifact, 0)
	module := c.QueryParam("module")
	artifactType := c.QueryParam("type")
	name := c.QueryParam("name")
	format := c.QueryParam("format")
	formatVersion := c.QueryParam("format_version")

	// filter artifacts
	for _, artifact := range hc.state.Artifacts {
		if len(module) > 0 && module != artifact.Module {
			continue
		}
		if len(artifactType) > 0 && artifactType != string(artifact.Type) {
			continue
		}
		if len(name) > 0 && name != artifact.Name {
			continue
		}
		if len(format) > 0 && format != artifact.Format {
			continue
		}
		if len(formatVersion) > 0 && formatVersion != artifact.FormatVersion {
			continue
		}

		result = append(result, artifact)
	}

	return c.JSON(http.StatusOK, result)
}

// artifactUpload uploads a report (typically from code scanning)
func (hc *handlerConfig) artifactUpload(c echo.Context) error {
	moduleSlug := c.FormValue("module")
	fileType := c.FormValue("type")
	format := c.FormValue("format")
	formatVersion := c.FormValue("format_version")
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// module is required, default to root
	if moduleSlug == "" {
		moduleSlug = "root"
	}

	// target dir
	targetDir := path.Join(hc.artifactDir, moduleSlug, fileType)
	_ = os.MkdirAll(targetDir, os.FileMode(0700))

	// store file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(path.Join(hc.artifactDir, moduleSlug, fileType, file.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// store into state
	hc.state.Artifacts[moduleSlug+"/"+file.Filename] = state.ActionArtifact{
		BuildID:       hc.buildID,
		JobID:         hc.jobID,
		Module:        moduleSlug,
		Type:          state.ActionArtifactType(fileType),
		Name:          file.Filename,
		Format:        format,
		FormatVersion: formatVersion,
	}

	return nil
}

// artifactDownload uploads a report (typically from code scanning)
func (hc *handlerConfig) artifactDownload(c echo.Context) error {
	moduleSlug := c.QueryParam("module")
	fileType := c.QueryParam("type")
	name := c.QueryParam("name")

	artifactFile := path.Join(hc.artifactDir, moduleSlug, fileType, name)
	return c.File(artifactFile)
}
