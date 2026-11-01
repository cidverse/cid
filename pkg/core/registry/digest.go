package registry

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
	orasregistry "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// ResolveReferenceDigest resolves an image reference (tag or digest) to its manifest digest.
// Unlike GetArtifactDigest it uses a manifest HEAD request instead of downloading the manifest,
// which makes it cheap enough to resolve every tag of a repository.
func ResolveReferenceDigest(reference string) (string, error) {
	ref, err := orasregistry.ParseReference(reference)
	if err != nil {
		return "", fmt.Errorf("invalid image reference %s: %w", reference, err)
	}

	repo, err := remote.NewRepository(ref.Registry + "/" + ref.Repository)
	if err != nil {
		return "", fmt.Errorf("invalid image repository %s: %w", reference, err)
	}
	repo.Client = NewAuthClient()

	desc, err := repo.Resolve(context.Background(), ref.Reference)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", reference, err)
	}

	return desc.Digest.String(), nil
}

func GetArtifactDigest(reference string) (string, error) {
	// repo
	repo, _ := remote.NewRepository(reference)

	// fetch manifest
	fetchOpts := oras.DefaultFetchBytesOptions
	desc, _, err := oras.FetchBytes(context.Background(), repo, reference, fetchOpts)
	if err != nil {
		return "", err
	}

	return desc.Digest.String(), nil
}
