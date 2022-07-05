package container

import (
	"errors"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func LoadRemoteImageInformation(reference string) (*remote.Descriptor, error) {
	/*
		basicAuthn := &authn.Basic{
			Username: os.Getenv("DOCKER_USERNAME"),
			Password: os.Getenv("DOCKER_PASSWORD"),
		}
		withAuthOption := remote.WithAuth(basicAuthn)
	*/
	options := []remote.Option{}

	ref, err := name.ParseReference(reference)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse reference of the image %s , detail: %v", reference, err))
	}
	descriptor, err := remote.Get(ref, options...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get image %s , detail: %v", reference, err))
	}

	return descriptor, nil
}
