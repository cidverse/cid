package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func PushCatalog(ref string, registryHost string, username string, password string, catalogFile string) (*ocispec.Descriptor, error) {
	ctx := context.Background()
	refParts := strings.SplitN(ref, ":", 2)

	// 0. Create a file store
	fs, err := file.New("/tmp/")
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	// add files to file store
	fileNames := []string{catalogFile}
	fileDescriptors := make([]ocispec.Descriptor, 0, len(fileNames))
	for _, name := range fileNames {
		fileDescriptor, addErr := fs.Add(ctx, "cid-index.yaml", OCICatalogFileMediaType, name)
		if addErr != nil {
			return nil, addErr
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
	}

	// manifest
	var manifestAnnotations = map[string]string{
		ocispec.AnnotationCreated: time.Now().UTC().Format(time.RFC3339),
	}
	manifestDescriptor, err := oras.Pack(ctx, fs, OCICatalogManifestMediaType, fileDescriptors, oras.PackOptions{
		PackImageManifest:   true,
		ManifestAnnotations: manifestAnnotations,
	})
	if err != nil {
		return nil, err
	}

	// tag
	if err = fs.Tag(ctx, manifestDescriptor, refParts[1]); err != nil {
		return nil, err
	}

	// remote repository
	repo, err := remote.NewRepository(refParts[0])
	if err != nil {
		return nil, err
	}
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(registryHost, auth.Credential{
			Username: username,
			Password: password,
		}),
	}

	// copy from file store to the remote
	mf, err := oras.Copy(ctx, fs, refParts[1], repo, refParts[1], oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to push manifest: %w", err)
	}

	return &mf, err
}
