package gitlabreleasepublish

import (
	"bytes"
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"net/http"
	"os"
	"strconv"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cidverseutils/core/ci"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const URI = "builtin://actions/gitlab-release-publish"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	CIJobToken  string `json:"ci_job_token"  env:"CI_JOB_TOKEN"`
	GitLabToken string `json:"gitlab_token"  env:"GITLAB_TOKEN"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "gitlab-release-publish",
		Description: "Publishes a new release on GitLab, including the release notes and artifacts.",
		Category:    "publish",
		Scope:       cidsdk.ActionScopeProject,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `hasPrefix(ENV["NCI_REPOSITORY_REMOTE"], "https://gitlab.com/")`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
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
			Executables: []cidsdk.ActionAccessExecutable{},
		},
		Input: cidsdk.ActionInput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type: "changelog",
				},
				{
					Type: "binary",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ProjectActionData) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectActionDataV1()
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
	changelogFile := cidsdk.JoinPath(d.Config.TempDir, "gitlab.changelog")
	changelogErr := a.Sdk.ArtifactDownload(cidsdk.ArtifactDownloadRequest{
		ID:         "root|changelog|gitlab.changelog",
		TargetFile: changelogFile,
	})
	if changelogErr != nil {
		releaseNotes.WriteString(fmt.Sprintf("No changelog available.\n"))
	} else {
		content, err := a.Sdk.FileRead(changelogFile)
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
	artifacts, err := a.Sdk.ArtifactList(cidsdk.ArtifactListRequest{Query: `artifact_type == "binary"`})
	if err != nil {
		return err
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "searching for artifacts to include in the release", Context: map[string]interface{}{"artifact_count": len(*artifacts)}})
	for _, artifact := range *artifacts {
		targetFile := cidsdk.JoinPath(d.Config.TempDir, artifact.Name)
		var dlErr = a.Sdk.ArtifactDownload(cidsdk.ArtifactDownloadRequest{
			ID:         artifact.ID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "error", Message: "failed to retrieve release artifact", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, artifact.Name)}})
			return fmt.Errorf("failed to retrieve release artifact: %w", dlErr)
		}

		reader, err := os.Open(targetFile)
		if err != nil {
			_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "warn", Message: "failed to open release artifact", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, artifact.Name)}})
			continue
		}
		defer reader.Close()
		gpf, _, err := glab.GenericPackages.PublishPackageFile(projectId, artifact.Module, releaseVersion, artifact.Name, reader, &gitlab.PublishPackageFileOptions{
			Status: gitlab.Ptr(gitlab.PackageDefault),
			Select: gitlab.Ptr(gitlab.SelectPackageFile),
		})
		if err != nil {
			return fmt.Errorf("failed to upload package: %w", err)
		}

		assetUrl := fmt.Sprintf("%s/-/package_files/%d/download", d.Env["NCI_PROJECT_URL"], gpf.ID)
		releaseAssets = append(releaseAssets, &gitlab.ReleaseAssetLinkOptions{
			Name:     gitlab.Ptr(fmt.Sprintf("%s/%s", artifact.Module, gpf.FileName)),
			URL:      gitlab.Ptr(assetUrl),
			LinkType: gitlab.Ptr(gitlab.OtherLinkType),
		})
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "added artifact to release", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, artifact.Name), "url": assetUrl}})
	}

	// create release
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "creating release", Context: map[string]interface{}{"gl_project_id": projectId, "gl_host": host}})
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
