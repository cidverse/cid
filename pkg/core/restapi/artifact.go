package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/core/expression"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cid/pkg/core/util"
	"github.com/cidverse/cidverseutils/pkg/archive/tar"
	"github.com/cidverse/cidverseutils/pkg/archive/zip"
	"github.com/cidverse/cidverseutils/pkg/encoding"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
)

// artifactList lists all generated reports
func (hc *APIConfig) artifactList(c echo.Context) error {
	var result = make([]state.ActionArtifact, 0)

	// query expression
	expr := util.GetStringOrDefault(c.QueryParam("query"), "true")

	// filter artifacts
	log.Debug().Str("query", expr).Msg("querying artifact list")
	for _, artifact := range hc.State.Artifacts {
		add, err := expression.EvalBooleanExpression(expr, map[string]interface{}{
			"id":             artifact.ArtifactID,
			"module":         artifact.Module,
			"artifact_type":  artifact.Type,
			"name":           artifact.Name,
			"format":         artifact.Format,
			"format_version": artifact.FormatVersion,
		})
		if err != nil {
			return fmt.Errorf("failed to evaluate query [%s]: %w", expr, err)
		}

		if add {
			result = append(result, artifact)
		}
	}

	return c.JSON(http.StatusOK, result)
}

// artifactUpload uploads a report (typically from code scanning)
func (hc *APIConfig) artifactUpload(c echo.Context) error {
	moduleSlug := util.GetStringOrDefault(c.FormValue("module"), "root")
	fileType := c.FormValue("type")
	format := c.FormValue("format")
	formatVersion := c.FormValue("format_version")
	extractFile := util.GetStringOrDefault(c.FormValue("extract_file"), "false")
	extractFileBool, _ := strconv.ParseBool(extractFile)
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

	// store
	fileHash, err := hc.storeArtifact(moduleSlug, fileType, format, formatVersion, file.Filename, src, extractFileBool)
	if err != nil {
		return err
	}

	// generate build provenance?
	if funk.Contains(provenance.FileTypes, fileType) {
		log.Info().Str("artifact", file.Filename).Str("type", fileType).Msg("generating provenance for artifact")
		prov := provenance.GenerateInTotoPredicate(file.Filename, fileHash, hc.Env, hc.State)

		provJSON, provErr := json.Marshal(prov)
		if provErr != nil {
			return provErr
		}

		_, err = hc.storeArtifact(moduleSlug, "attestation", "provenance", v1.PredicateSLSAProvenance, file.Filename, bytes.NewReader(provJSON), false)
		if err != nil {
			return err
		}
	}

	return nil
}

// artifactDownload uploads a report (typically from code scanning)
func (hc *APIConfig) artifactDownload(c echo.Context) error {
	id := c.QueryParam("id")
	moduleSlug := util.GetStringOrDefault(c.FormValue("module"), "root")
	fileType := c.QueryParam("type")
	name := c.QueryParam("name")

	// if set, use id
	if len(id) > 0 {
		parts := strings.SplitN(id, "|", 3)
		moduleSlug = parts[0]
		fileType = parts[1]
		name = parts[2]
	}

	artifactFile := path.Join(hc.ArtifactDir, moduleSlug, fileType, name)
	return c.File(artifactFile)
}

// storeArtifact stores an artifact on the local filesystem
func (hc *APIConfig) storeArtifact(moduleSlug string, fileType string, format string, formatVersion string, name string, content io.Reader, extract bool) (string, error) {
	var hashReader bytes.Buffer
	contentReader := io.TeeReader(content, &hashReader)

	// target dir
	targetDir := path.Join(hc.ArtifactDir, moduleSlug, fileType)
	targetFile := path.Join(targetDir, name)
	_ = os.MkdirAll(targetDir, os.ModePerm)

	// store file
	dst, err := os.Create(targetFile)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, contentReader); err != nil {
		return "", err
	}

	// sha256 hash
	fileHash, err := encoding.SHA256Hash(&hashReader)
	if err != nil {
		return "", err
	}

	// store into state
	hc.State.Artifacts[moduleSlug+"/"+name] = state.ActionArtifact{
		BuildID:       hc.BuildID,
		JobID:         hc.JobID,
		ArtifactID:    fmt.Sprintf("%s|%s|%s", moduleSlug, fileType, name),
		Module:        moduleSlug,
		Type:          fileType,
		Name:          name,
		Format:        format,
		FormatVersion: formatVersion,
		SHA256:        fileHash,
	}

	// allow to extract assets (github pages, gitlab pages, etc.)
	if extract {
		extractTargetDir := path.Join(targetDir, strings.TrimSuffix(name, filepath.Ext(name)))
		_ = os.MkdirAll(extractTargetDir, os.ModePerm)

		log.Debug().Str("target_dir", extractTargetDir).Str("format", format).Msg("extracting artifact archive")
		if format == "tar" {
			err := tar.Extract(targetFile, extractTargetDir)
			if err != nil {
				return "", err
			}
		} else if format == "zip" {
			err := zip.Extract(targetFile, extractTargetDir)
			if err != nil {
				return "", err
			}
		}
	}

	return fileHash, nil
}
