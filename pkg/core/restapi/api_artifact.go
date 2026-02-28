package restapi

import (
	"fmt"
	"net/http"

	"github.com/cidverse/cid/pkg/core/actionsdk"

	"github.com/cidverse/cid/pkg/util"
	"github.com/labstack/echo/v5"
)

func (hc *APIConfig) artifactList(c *echo.Context) error {
	// parameters
	expression := util.GetStringOrDefault(c.QueryParam("query"), "true")

	// query
	response, err := hc.SDKClient.ArtifactListV1(actionsdk.ArtifactListRequest{
		Query: expression,
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to get module list", err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

func (hc *APIConfig) artifactUpload(c *echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// reader
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// bytes
	fileBytes := make([]byte, file.Size)
	_, err = src.Read(fileBytes)
	if err != nil {
		return err
	}

	// store
	_, _, err = hc.SDKClient.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:          file.Filename,
		ContentBytes:  fileBytes,
		Module:        util.GetStringOrDefault(c.FormValue("module"), "root"),
		Type:          c.FormValue("type"),
		Format:        c.FormValue("format"),
		FormatVersion: c.FormValue("format_version"),
		ExtractFile:   util.GetStringOrDefault(c.FormValue("extract_file"), "false") == "true",
	})
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, "failed to upload artifact", err.Error())
	}

	return nil
}

func (hc *APIConfig) artifactDownload(c *echo.Context) error {
	artifact, err := hc.SDKClient.ArtifactByIdV1(c.QueryParam("id"))
	if err != nil {
		return hc.handleError(c, http.StatusInternalServerError, fmt.Sprintf("artifact with id [%s] not found", c.QueryParam("id")), err.Error())
	}

	return c.File(artifact.Path)
}
