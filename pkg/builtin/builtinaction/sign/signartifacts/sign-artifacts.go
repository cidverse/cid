package signartifacts

import (
	"fmt"
	"os"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/lib/secret"
)

const URI = "builtin://actions/sign-artifacts"

type Action struct {
	Sdk actionsdk.SDKClient
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "sign-artifacts",
		Description: "Signs binary artifacts with an OpenPGP key and stores the detached signatures as additional artifacts.",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeProject,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `getMapValue(ENV, "GPG_PRIVATE_KEY") != ""`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{
				{
					Name:        "GPG_PRIVATE_KEY",
					Description: "The OpenPGP private key used to sign the artifacts. Signing is skipped when not set.",
					Required:    false,
					Secret:      true,
				},
				{
					Name:        "GPG_PASSPHRASE",
					Description: "The passphrase of the OpenPGP private key.",
					Required:    false,
					Secret:      true,
				},
			},
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
					Type:   "signature",
					Format: "pgp",
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

	// signing is optional - the plan rule already excludes this step when no key is configured,
	// this is the runtime safety net for plans generated with a key that is no longer available.
	if d.Env["GPG_PRIVATE_KEY"] == "" {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "no GPG_PRIVATE_KEY configured, skipping artifact signing"})
		return nil
	}

	// sign all binary artifacts
	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: `artifact_type == "binary"`})
	if err != nil {
		return err
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signing artifacts", Context: map[string]interface{}{"artifact_count": len(artifacts)}})
	for _, artifact := range artifacts {
		targetFile := actionsdk.JoinPath(d.Config.TempDir, artifact.Name)
		_, dlErr := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			return fmt.Errorf("failed to retrieve artifact %s: %w", artifact.ArtifactID, dlErr)
		}

		content, readErr := os.ReadFile(targetFile)
		if readErr != nil {
			return fmt.Errorf("failed to read artifact %s: %w", artifact.ArtifactID, readErr)
		}

		signature, signErr := secret.SignOpenPGP(d.Env["GPG_PRIVATE_KEY"], d.Env["GPG_PASSPHRASE"], content)
		if signErr != nil {
			return fmt.Errorf("failed to sign artifact %s: %w", artifact.ArtifactID, signErr)
		}

		_, _, uploadErr := a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:         targetFile + ".asc",
			ContentBytes: []byte(signature),
			Module:       artifact.Module,
			Type:         "signature",
			Format:       "pgp",
		})
		if uploadErr != nil {
			return fmt.Errorf("failed to store signature for %s: %w", artifact.ArtifactID, uploadErr)
		}

		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "signed artifact", Context: map[string]interface{}{"artifact": artifact.Name, "signature": artifact.Name + ".asc"}})
	}

	return nil
}
