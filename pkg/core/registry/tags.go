package registry

import (
	"context"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type ImageTag struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
}

func FindTags(repositoryURL string) ([]ImageTag, error) {
	ctx := context.Background()

	// query tags
	repo, err := remote.NewRepository(repositoryURL)
	if err != nil {
		return []ImageTag{}, err
	}

	tagList, err := registry.Tags(ctx, repo)
	if err != nil {
		return []ImageTag{}, err
	}

	// add tags to list
	tags := make([]ImageTag, 0, len(tagList))
	for _, tag := range tagList {
		tags = append(tags, ImageTag{Repository: repositoryURL, Tag: tag, Digest: ""})
	}

	return tags, nil
}
