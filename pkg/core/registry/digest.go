package registry

import (
	"context"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

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
