package builtin

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cid/pkg/lib/formats/cobertura"
	"github.com/cidverse/cid/pkg/lib/formats/jacoco"
	"github.com/cidverse/cid/pkg/lib/githublib"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/compress"
	"github.com/cidverse/go-rules/pkg/expr"
	v1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
	"github.com/rs/zerolog/log"
)

func (sdk ActionSDK) ArtifactListV1(request actionsdk.ArtifactListRequest) ([]*actionsdk.Artifact, error) {
	// parameters
	expression := util.GetStringOrDefault(request.Query, "true")
	log.Debug().Str("query", expression).Msg("[API] artifact list query")

	// filter artifacts
	var result = make([]*actionsdk.Artifact, 0)
	for _, artifact := range sdk.State.Artifacts {
		add, err := expr.EvalBooleanExpression(expression, map[string]interface{}{
			"id":             artifact.ArtifactID,
			"module":         artifact.Module,
			"artifact_type":  artifact.Type,
			"name":           artifact.Name,
			"format":         artifact.Format,
			"format_version": artifact.FormatVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate query [%s]: %w", expression, err)
		}

		if add {
			result = append(result, &actionsdk.Artifact{
				BuildID:       artifact.BuildID,
				JobID:         artifact.JobID,
				StepSlug:      artifact.StepSlug,
				ArtifactID:    artifact.ArtifactID,
				Module:        artifact.Module,
				Type:          artifact.Type,
				Name:          artifact.Name,
				Path:          artifact.Path,
				Format:        artifact.Format,
				FormatVersion: artifact.FormatVersion,
				SHA256:        artifact.SHA256,
			})
		}
	}

	return result, nil
}

func (sdk ActionSDK) ArtifactByIdV1(id string) (*actionsdk.Artifact, error) {
	artifact, present := sdk.State.Artifacts[id]
	if !present {
		return nil, fmt.Errorf("artifact %s not present", id)
	}

	return &actionsdk.Artifact{
		BuildID:       artifact.BuildID,
		JobID:         artifact.JobID,
		StepSlug:      artifact.StepSlug,
		ArtifactID:    artifact.ArtifactID,
		Module:        artifact.Module,
		Type:          artifact.Type,
		Name:          artifact.Name,
		Path:          artifact.Path,
		Format:        artifact.Format,
		FormatVersion: artifact.FormatVersion,
		SHA256:        artifact.SHA256,
	}, nil
}

func (sdk ActionSDK) ArtifactUploadV1(request actionsdk.ArtifactUploadRequest) (filePath string, fileHash string, err error) {
	slog.With("file", request.File).With("type", request.Type).With("format", request.Format).With("format_version", request.FormatVersion).Info("storing artifact")

	// store
	targetFile, fileHash, err := sdk.storeArtifact(request)
	if err != nil {
		return "", "", err
	}

	// generate build provenance?
	if slices.Contains(provenance.FileTypes, request.Type) {
		log.Info().Str("artifact", request.File).Str("type", request.Type).Msg("generating provenance for artifact")
		prov := provenance.GenerateInTotoPredicate(request.File, fileHash, sdk.Env, sdk.State)

		provJSON, provErr := json.Marshal(prov)
		if provErr != nil {
			return "", "", provErr
		}

		_, _, err = sdk.storeArtifact(actionsdk.ArtifactUploadRequest{
			File:          fmt.Sprintf("%s-provenance.json", strings.TrimSuffix(request.File, filepath.Ext(request.File))),
			Module:        request.ModuleSlug(),
			Type:          "attestation",
			Format:        "provenance",
			FormatVersion: v1.PredicateSLSAProvenance,
			ContentBytes:  provJSON,
		})
		if err != nil {
			return "", "", err
		}
	}

	// post-process artifacts
	err = postProcessArtifact(&sdk, targetFile, request.Type, request.Format, request.FormatVersion)
	if err != nil {
		slog.With("err", err).Warn("failed to post-process artifact")
	}

	return targetFile, fileHash, nil
}

func (sdk ActionSDK) storeArtifact(request actionsdk.ArtifactUploadRequest) (filePath string, fileHash string, err error) {
	moduleSlug := util.GetStringOrDefault(request.Module, "root")

	var reader io.Reader
	if request.Content != "" {
		reader = strings.NewReader(request.Content)
	} else if request.ContentBytes != nil && len(request.ContentBytes) > 0 {
		reader = bytes.NewReader(request.ContentBytes)
	} else if request.File != "" {
		// security
		if err = util.ValidateFileInAllowedDirs(request.File, sdk.ProjectDir, sdk.TempDir); err != nil {
			return "", "", fmt.Errorf("file %s is not located in allowed directories: %w", request.File, err)
		}

		// open file
		file, err := os.Open(request.File)
		if err != nil {
			return "", "", err
		}

		reader = file
		defer file.Close()
	} else {
		return "", "", fmt.Errorf("file, content or contentBytes must be provided when uploading an artifact")
	}

	// target dir
	targetDir := path.Join(sdk.ArtifactDir, sdk.Step.Slug, request.Type)
	if request.Format != "" {
		targetDir = path.Join(targetDir, request.Format)
	}
	fileName := filepath.Base(request.File)
	targetFile := path.Join(targetDir, fileName)
	_ = os.MkdirAll(targetDir, os.ModePerm)

	// create file
	dst, err := os.Create(targetFile)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	// write and hash content
	hasher := sha256.New()
	tee := io.TeeReader(reader, hasher)
	if _, err = io.Copy(dst, tee); err != nil {
		return "", "", err
	}
	if err = dst.Sync(); err != nil {
		return "", "", err
	}
	fileHash = hex.EncodeToString(hasher.Sum(nil))

	// store into state
	slog.With("module", moduleSlug).With("type", request.Type).With("format", request.Format).With("format_version", request.FormatVersion).With("file", targetFile).With("hash", fileHash).Info("[API] action output artifact stored")
	sdk.State.Artifacts[fmt.Sprintf("%s|%s|%s", moduleSlug, request.Type, request.File)] = state.ActionArtifact{
		BuildID:       sdk.BuildID,
		JobID:         sdk.JobID,
		StepSlug:      sdk.Step.Slug,
		ArtifactID:    fmt.Sprintf("%s|%s|%s", moduleSlug, request.Type, request.File),
		Module:        moduleSlug,
		Type:          request.Type,
		Name:          request.File,
		Path:          targetFile,
		Format:        request.Format,
		FormatVersion: request.FormatVersion,
		SHA256:        fileHash,
	}

	// allow to extract assets (github pages, gitlab pages, etc.)
	if request.ExtractFile {
		extractTargetDir := path.Join(targetDir, strings.TrimSuffix(request.File, filepath.Ext(request.File)))
		_ = os.MkdirAll(extractTargetDir, os.ModePerm)

		log.Debug().Str("target_dir", extractTargetDir).Str("format", request.Format).Msg("extracting artifact archive")
		if request.Format == "tar" {
			err = compress.TARExtract(targetFile, extractTargetDir)
			if err != nil {
				return "", "", err
			}
		} else if request.Format == "zip" {
			err = compress.ZIPExtract(targetFile, extractTargetDir)
			if err != nil {
				return "", "", err
			}
		}
	}

	return targetFile, fileHash, nil
}

func (sdk ActionSDK) ArtifactDownloadV1(request actionsdk.ArtifactDownloadRequest) (*actionsdk.ArtifactDownloadResult, error) {
	if request.ID == "" {
		return nil, fmt.Errorf("id must be provided")
	}
	if request.TargetFile == "" {
		return nil, fmt.Errorf("target_file must be provided")
	}

	id := request.ID
	slog.With("id", id).Info("[API] artifact download requested")

	// lookup
	artifact, present := sdk.State.Artifacts[id]
	if !present {
		return nil, fmt.Errorf("artifact %s not present", id)
	}

	_ = os.MkdirAll(path.Dir(request.TargetFile), os.ModePerm)
	src, err := os.Open(artifact.Path)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	dst, err := os.Create(request.TargetFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = dst.Close()
	}()

	// copy + hash simultaneously
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)
	size, err := io.Copy(dst, tee)
	if err != nil {
		return nil, err
	}

	// disk flush
	if err = dst.Sync(); err != nil {
		return nil, err
	}

	return &actionsdk.ArtifactDownloadResult{
		Path: request.TargetFile,
		Hash: hex.EncodeToString(hasher.Sum(nil)),
		Size: size,
	}, nil
}

func (sdk ActionSDK) ArtifactDownloadByteArrayV1(request actionsdk.ArtifactDownloadByteArrayRequest) (*actionsdk.ArtifactDownloadByteArrayResult, error) {
	if request.ID == "" {
		return nil, fmt.Errorf("id must be provided")
	}

	id := request.ID
	slog.With("id", id).Info("[API] artifact download requested")

	// lookup
	artifact, present := sdk.State.Artifacts[id]
	if !present {
		return nil, fmt.Errorf("artifact %s not present", id)
	}

	// open file
	src, err := os.Open(artifact.Path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// copy + hash simultaneously
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)
	data, err := io.ReadAll(tee)
	if err != nil {
		return nil, err
	}

	return &actionsdk.ArtifactDownloadByteArrayResult{
		Bytes: data,
		Hash:  hex.EncodeToString(hasher.Sum(nil)),
		Size:  int64(len(data)),
	}, nil
}

func postProcessArtifact(sdk *ActionSDK, targetFile string, fileType string, format string, formatVersion string) error {
	switch {
	case fileType == "report" && format == "jacoco":
		coverage, err := jacoco.ParseCoverageFromFile(targetFile)
		if err != nil {
			return fmt.Errorf("failed to parse jacoco report: %w", err)
		}

		slog.With("file", targetFile).With("coverage", coverage).Info("[API] calculated coverage from jacoco report")
		if sdk.NCI.Repository.HostType == "gitlab" {
			fmt.Printf("Test-Coverage:%.2f%%\n", coverage) // some platforms parse the test-coverage from stdout (e.g. GitLab)
		}

	case fileType == "report" && format == "cobertura":
		coverage, err := cobertura.ParseCoverageFromFile(targetFile)
		if err != nil {
			return fmt.Errorf("failed to parse cobertura report: %w", err)
		}

		slog.With("file", targetFile).With("coverage", coverage).Info("[API] calculated coverage from cobertura report")
		if sdk.NCI.Repository.HostType == "gitlab" {
			fmt.Printf("Test-Coverage:%.2f%%\n", coverage) // some platforms parse the test-coverage from stdout (e.g. GitLab)
		}

	case fileType == "report" && format == "sarif" && formatVersion == "2.1.0" && sdk.NCI.Repository.HostType == "gitlab": // TD-001: automatic conversion of SARIF to GitLab Code Quality due to missing SARIF support in GitLab
		// temp file
		codeQualityFile, err := os.CreateTemp(os.TempDir(), "gl-code-quality-report-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer codeQualityFile.Close() // Ensure the temporary file is closed when done

		// code-quality report
		cmdResult, err := sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
			WorkDir: sdk.ProjectDir,
			Command: fmt.Sprintf("gitlab-sarif-converter --type=codequality %q %q", targetFile, codeQualityFile.Name()),
		})
		if err != nil {
			return fmt.Errorf("failed to execute gitlab-sarif-converter: %w", err)
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("gitlab-sarif-converter failed, exit code %d: %s", cmdResult.Code, cmdResult.Error)
		}

		// upload
		moduleSlug := ""
		if sdk.CurrentModule != nil {
			moduleSlug = sdk.CurrentModule.Slug
		}

		_, _, err = sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:          codeQualityFile.Name(),
			Content:       "",
			ContentBytes:  nil,
			Module:        moduleSlug,
			Type:          "report",
			Format:        "gl-codequality",
			FormatVersion: "",
		})
		if err != nil {
			return fmt.Errorf("failed to store converted code quality report: %w", err)
		}

	case fileType == "report" && format == "sarif" && formatVersion == "2.1.0" && sdk.NCI.Repository.HostType == "github":
		err := githublib.GitHubCodeSecuritySarifUpload(os.Getenv("GITHUB_TOKEN"), targetFile, sdk.NCI)
		if err != nil {
			return fmt.Errorf("failed to upload sarif report to github: %w", err)
		}
	}

	return nil
}
