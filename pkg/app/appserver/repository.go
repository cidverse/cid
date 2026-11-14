package appserver

import (
	"errors"
	"fmt"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

// errRepositoryNotAccessible marks repositories that exist but must not be rendered (same rules as the app's repository listing).
var errRepositoryNotAccessible = errors.New("repository not accessible")

// lookupRepository fetches the repository live from the platform.
// FindRepositoryById returns everything the token can read (e.g. any public project on gitlab.com),
// so the filters of Platform.Repositories (maintainer access, not archived) are re-applied here.
func (s *Server) lookupRepository(id int64) (api.Repository, error) {
	repo, err := s.cfg.Platform.FindRepositoryById(id)
	if err != nil {
		return api.Repository{}, err
	}

	if repo.IsArchived {
		return api.Repository{}, fmt.Errorf("%w: %s is archived", errRepositoryNotAccessible, repo.Path)
	}
	if repo.IsEmpty {
		return api.Repository{}, fmt.Errorf("%w: %s is empty", errRepositoryNotAccessible, repo.Path)
	}
	if project, ok := repo.InternalRepo.(*gitlab.Project); ok && !hasGitLabMaintainerAccess(project) {
		return api.Repository{}, fmt.Errorf("%w: %s requires maintainer access", errRepositoryNotAccessible, repo.Path)
	}

	return repo, nil
}

// hasGitLabMaintainerAccess mirrors the MinAccessLevel=Maintainer filter used when listing GitLab projects.
// group_access reflects the highest membership in the project's group hierarchy.
func hasGitLabMaintainerAccess(project *gitlab.Project) bool {
	if project.Permissions == nil {
		return false
	}
	if pa := project.Permissions.ProjectAccess; pa != nil && pa.AccessLevel >= gitlab.MaintainerPermissions {
		return true
	}
	if ga := project.Permissions.GroupAccess; ga != nil && ga.AccessLevel >= gitlab.MaintainerPermissions {
		return true
	}

	return false
}
