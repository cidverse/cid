package checksumartifacts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/checksum-artifacts"

type Action struct {
	Sdk actionsdk.SDKClient
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "checksum-artifacts",
		Description: "Creates SHA-256 checksums for binary artifacts: one <asset>.sha256 file per binary and an aggregated SHA256SUMS file.",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeProject,
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{},
		},
		Input: actionsdk.ActionInput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type: "binary",
				},
			},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   common.ArtifactTypeChecksum,
					Format: common.ArtifactFormatSHA256,
				},
			},
		},
	}
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectExecutionContextV1()
	if err != nil {
		return err
	}

	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`})
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "no binary artifacts found, skipping checksum generation"})
		return nil
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "creating checksums", Context: map[string]interface{}{"artifact_count": len(artifacts)}})

	assetNames := common.ReleaseAssetNames(artifacts)
	var lines []string
	for _, artifact := range artifacts {
		assetName := assetNames[artifact.ArtifactID]
		targetFile := actionsdk.JoinPath(d.Config.TempDir, assetName)
		_, dlErr := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			return fmt.Errorf("failed to retrieve artifact %s: %w", artifact.ArtifactID, dlErr)
		}

		digest, sumErr := common.SHA256File(targetFile)
		if sumErr != nil {
			return fmt.Errorf("failed to create checksum for %s: %w", artifact.ArtifactID, sumErr)
		}
		line := common.ChecksumLine(digest, assetName)
		lines = append(lines, line)

		_, _, uploadErr := a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:         targetFile + ".sha256",
			ContentBytes: []byte(line),
			Module:       artifact.Module,
			Type:         common.ArtifactTypeChecksum,
			Format:       common.ArtifactFormatSHA256,
		})
		if uploadErr != nil {
			return fmt.Errorf("failed to store checksum for %s: %w", artifact.ArtifactID, uploadErr)
		}
	}

	// aggregated checksum file, sorted by asset name for a stable output
	sort.Slice(lines, func(i, j int) bool {
		return strings.SplitN(lines[i], "  ", 2)[1] < strings.SplitN(lines[j], "  ", 2)[1]
	})
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:         actionsdk.JoinPath(d.Config.TempDir, common.ChecksumsFileName),
		ContentBytes: []byte(strings.Join(lines, "")),
		Module:       common.ChecksumsModule,
		Type:         common.ArtifactTypeChecksum,
		Format:       common.ArtifactFormatSHA256,
	})
	if err != nil {
		return fmt.Errorf("failed to store %s: %w", common.ChecksumsFileName, err)
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "created checksums", Context: map[string]interface{}{"file": common.ChecksumsFileName, "entries": len(lines)}})

	return nil
}
