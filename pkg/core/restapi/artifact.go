package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/PhilippHeuer/in-toto-golang/in_toto/slsa_provenance/v1.0"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/pkg/encoding"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/thoas/go-funk"
)

// artifactList lists all generated reports
func (hc *APIConfig) artifactList(c echo.Context) error {
	var result = make([]state.ActionArtifact, 0)
	module := c.QueryParam("module")
	artifactType := c.QueryParam("type")
	name := c.QueryParam("name")
	format := c.QueryParam("format")
	formatVersion := c.QueryParam("format_version")

	// filter artifacts
	for _, artifact := range hc.State.Artifacts {
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
func (hc *APIConfig) artifactUpload(c echo.Context) error {
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

	// reader
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// store
	fileHash, err := hc.storeArtifact(moduleSlug, fileType, format, formatVersion, file.Filename, src)
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

		_, err = hc.storeArtifact(moduleSlug, "attestation", "provenance", v1.PredicateSLSAProvenance, file.Filename, bytes.NewReader(provJSON))
		if err != nil {
			return err
		}
	}

	return nil
}

// artifactDownload uploads a report (typically from code scanning)
func (hc *APIConfig) artifactDownload(c echo.Context) error {
	id := c.QueryParam("id")
	moduleSlug := c.QueryParam("module")
	fileType := c.QueryParam("type")
	name := c.QueryParam("name")

	// module is required, default to root
	if moduleSlug == "" {
		moduleSlug = "root"
	}

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
func (hc *APIConfig) storeArtifact(moduleSlug string, fileType string, format string, formatVersion string, name string, content io.Reader) (string, error) {
	var hashReader bytes.Buffer
	contentReader := io.TeeReader(content, &hashReader)

	// target dir
	targetDir := path.Join(hc.ArtifactDir, moduleSlug, fileType)
	_ = os.MkdirAll(targetDir, os.ModePerm)

	// store file
	dst, err := os.Create(path.Join(hc.ArtifactDir, moduleSlug, fileType, name))
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
		Type:          state.ActionArtifactType(fileType),
		Name:          name,
		Format:        format,
		FormatVersion: formatVersion,
		SHA256:        fileHash,
	}

	return fileHash, nil
}
