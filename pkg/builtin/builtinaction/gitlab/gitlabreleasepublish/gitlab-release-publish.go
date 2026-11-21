package gitlabreleasepublish

import (
	"bytes"
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"net/http"
	"os"
	"strconv"

	"github.com/cidverse/cidverseutils/core/ci"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

const URI = "builtin://actions/gitlab-release-publish"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	CIJobToken  string `json:"ci_job_token"  env:"CI_JOB_TOKEN"`
	GitLabToken string `json:"gitlab_token"  env:"GITLAB_TOKEN"`
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "gitlab-release-publish",
		Description: "Publishes a new release on GitLab, including the release notes and artifacts.",
		Category:    "publish",
		Scope:       actionsdk.ActionScopeProject,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `hasPrefix(ENV["NCI_REPOSITORY_REMOTE"], "https://gitlab.com/")`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{
				{
					Name:        "CI_JOB_TOKEN",
					Description: "The GitLab job token to use for authentication. Preferred over personal access tokens.",
					Required:    false,
					Secret:      true,
				},
				{
					Name:        "GITLAB_TOKEN",
					Description: "The GitLab token to use for creating the release when using a personal access token.",
					Required:    false,
				},
				{
					Name:        "GITLAB_BOT_TOKEN",
					Description: "The GitLab token to use for creating the release when using a project bot account.",
					Required:    false,
				},
			},
			Executables: []actionsdk.ActionAccessExecutable{},
		},
		Input: actionsdk.ActionInput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type: "changelog",
				},
				{
					Type: "binary",
				},
				{
					Type:   "signature",
					Format: "pgp",
				},
				{
					Type:   "signature",
					Format: "sigstore",
				},
				{
					Type:   common.ArtifactTypeChecksum,
					Format: common.ArtifactFormatSHA256,
				},
				{
					Type:   common.SBOMArtifactType,
					Format: common.SBOMFormatCycloneDX,
				},
				{
					Type:   common.SBOMArtifactType,
					Format: common.SBOMFormatSPDX,
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ProjectExecutionContextV1Response) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// release notes / changelog
	var releaseNotes bytes.Buffer
	changelogFile := actionsdk.JoinPath(d.Config.TempDir, "gitlab.changelog")
	_, changelogErr := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|gitlab.changelog",
		TargetFile: changelogFile,
	})
	if changelogErr != nil {
		releaseNotes.WriteString(fmt.Sprintf("No changelog available.\n"))
	} else {
		content, err := a.Sdk.FileReadV1(changelogFile)
		if err != nil {
			releaseNotes.WriteString(fmt.Sprintf("No changelog available.\n"))
		} else {
			releaseNotes.WriteString(fmt.Sprintf("%s\n", content))
		}
	}

	// support for self-hosted instances
	host, err := ci.GetHostFromGitRemote(d.Env["NCI_REPOSITORY_REMOTE"])
	if err != nil {
		return err
	}

	// init client
	var glab *gitlab.Client
	if cfg.CIJobToken != "" {
		glab, err = gitlab.NewJobClient(cfg.CIJobToken, gitlab.WithBaseURL("https://"+host), gitlab.WithHTTPClient(&http.Client{Transport: http.DefaultTransport}))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
	} else {
		return fmt.Errorf("no available authentication method")
	}
	projectId, err := strconv.Atoi(d.Env["NCI_PROJECT_ID"])
	if err != nil {
		return fmt.Errorf("failed to parse project ID: %w", err)
	}
	releaseVersion := d.Env["NCI_COMMIT_REF_RELEASE"]

	// release artifacts
	var releaseAssets []*gitlab.ReleaseAssetLinkOptions
	artifacts, err := a.Sdk.ArtifactListV1(actionsdk.ArtifactListRequest{Query: common.ReleaseArtifactQuery})
	if err != nil {
		return err
	}
	artifacts = common.ReleaseAssets(artifacts)
	assetNames := common.ReleaseFileNames(artifacts)
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "searching for artifacts to include in the release", Context: map[string]interface{}{"artifact_count": len(artifacts)}})
	for _, artifact := range artifacts {
		// binaries, sboms (and their provenance) get their collision-free release asset name, signatures and checksums are already named after it
		assetName := assetNames[artifact.ArtifactID]
		targetFile := actionsdk.JoinPath(d.Config.TempDir, assetName)
		_, dlErr := a.Sdk.ArtifactDownloadV1(actionsdk.ArtifactDownloadRequest{
			ID:         artifact.ArtifactID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "error", Message: "failed to retrieve release artifact", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, artifact.Name)}})
			return fmt.Errorf("failed to retrieve release artifact: %w", dlErr)
		}

		reader, err := os.Open(targetFile)
		if err != nil {
			_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "failed to open release artifact", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, assetName)}})
			continue
		}
		gpf, _, err := glab.GenericPackages.PublishPackageFile(projectId, artifact.Module, releaseVersion, assetName, reader, &gitlab.PublishPackageFileOptions{
			Status: gitlab.Ptr(gitlab.PackageDefault),
			Select: gitlab.Ptr(gitlab.SelectPackageFile),
		})
		_ = reader.Close()
		if err != nil {
			return fmt.Errorf("failed to upload package: %w", err)
		}

		// asset names are already unique within the release (module prefix only on collisions), the generic package
		// per module is just storage - don't repeat the module in the link name
		assetUrl := fmt.Sprintf("%s/-/package_files/%d/download", d.Env["NCI_PROJECT_URL"], gpf.ID)
		releaseAssets = append(releaseAssets, &gitlab.ReleaseAssetLinkOptions{
			Name:     gitlab.Ptr(assetName),
			URL:      gitlab.Ptr(assetUrl),
			LinkType: gitlab.Ptr(gitlab.OtherLinkType),
		})
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "added artifact to release", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, assetName), "url": assetUrl}})
	}

	// create release
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "creating release", Context: map[string]interface{}{"gl_project_id": projectId, "gl_host": host}})
	_, _, err = glab.Releases.CreateRelease(projectId, &gitlab.CreateReleaseOptions{
		Name:        gitlab.Ptr(releaseVersion),               // release name
		TagName:     gitlab.Ptr(d.Env["NCI_COMMIT_REF_NAME"]), // tag name
		Description: gitlab.Ptr(releaseNotes.String()),        // release description
		Assets: &gitlab.ReleaseAssetsOptions{
			Links: releaseAssets,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	return nil
}
