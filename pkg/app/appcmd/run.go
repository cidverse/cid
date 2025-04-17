package appcmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appgithub"
	"github.com/cidverse/cid/pkg/app/appgitlab"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/task/taskcommon"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-run",
		Aliases: []string{},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			channel, _ := cmd.Flags().GetString("channel")

			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("Failed to configure platform from environment", "err", err)
				os.Exit(1)
			}

			// list repositories
			repos, err := platform.Repositories(api.RepositoryListOpts{
				IncludeBranches:   true,
				IncludeCommitHash: true,
			})
			if err != nil {
				slog.Error("Failed to list repositories", "err", err)
				os.Exit(1)
			}

			// execute task for each repository
			for _, repo := range repos {
				err = processRepository(platform, repo, channel)
				if err != nil {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).With("err", err).Warn("Failed to process repository")
				}
			}
		},
	}

	cmd.Flags().StringP("channel", "c", "", "Channel")

	return cmd
}

func processRepository(platform api.Platform, repo api.Repository, channel string) error {
	// only process repositories with a matching channel value
	if appcommon.GetChannel(platform, repo) != channel {
		slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to channel mismatch")
		return nil
	}

	// create temp directory
	tempDir, err := os.MkdirTemp("", "cid-app-*")
	if err != nil {
		return fmt.Errorf("failed to prepare temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// run platform-specific task
	slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("repo_channel", channel).With("platform", platform.Name()).Info("running workflow update task")
	taskContext := taskcommon.TaskContext{
		Directory:  filepath.Join(tempDir, slug.Make(repo.Name)),
		Platform:   platform,
		Repository: repo,
	}
	err = os.MkdirAll(taskContext.Directory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if platform.Slug() == "github" {
		return appgithub.GitHubWorkflowTask(taskContext)
	} else if platform.Slug() == "gitlab" {
		return appgitlab.GitLabWorkflowTask(taskContext)
	} else {
		return fmt.Errorf("platform %s not supported", platform.Slug())
	}
}
