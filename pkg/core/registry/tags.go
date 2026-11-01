package registry

import (
	"context"
	"log/slog"
	"sort"

	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type ImageTag struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest,omitempty"` // Digest is the manifest digest the tag resolved to, empty if unresolved
}

// FindTagsWithDigest returns every tag of a repository together with the manifest digest it currently
// resolves to. Tags that cannot be resolved are omitted: an unpinned tag must never be used to select
// an executable, because the tag can be moved to a different image after discovery.
// The result is sorted by tag, as the registry returns tags in an arbitrary order.
func FindTagsWithDigest(repositoryURL string) ([]ImageTag, error) {
	ctx := context.Background()

	repo, err := remote.NewRepository(repositoryURL)
	if err != nil {
		return nil, err
	}
	repo.Client = NewAuthClient()

	tagList, err := registry.Tags(ctx, repo)
	if err != nil {
		return nil, err
	}

	resolvePool := pool.NewWithResults[ImageTag]().WithMaxGoroutines(8)
	for _, tag := range tagList {
		t := tag
		resolvePool.Go(func() ImageTag {
			desc, resolveErr := repo.Resolve(ctx, t)
			if resolveErr != nil {
				slog.With("repository", repositoryURL).With("tag", t).With("err", resolveErr).Warn("failed to resolve tag digest, skipping tag")
				return ImageTag{}
			}

			return ImageTag{Repository: repositoryURL, Tag: t, Digest: desc.Digest.String()}
		})
	}

	resolved := make([]ImageTag, 0, len(tagList))
	for _, tag := range resolvePool.Wait() {
		if tag.Tag == "" || tag.Digest == "" {
			continue
		}
		resolved = append(resolved, tag)
	}
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Tag < resolved[j].Tag
	})

	return resolved, nil
}
