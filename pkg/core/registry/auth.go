package registry

import (
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// NewAuthClient returns an auth client that reuses the local container runtime credentials
// (e.g. ~/.docker/config.json), falling back to anonymous access when no credential store is available.
func NewAuthClient() *auth.Client {
	client := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}

	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err == nil {
		client.Credential = credentials.Credential(store)
	}

	return client
}
