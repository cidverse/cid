package storage

import (
	"os"

	"github.com/cidverse/cid/pkg/lib/storage/storageapi"
	"github.com/cidverse/cid/pkg/lib/storage/storages3"
)

func GetStorageApi() (storageapi.API, error) {
	s3Endpoint := os.Getenv("CID_STORAGE_S3_ENDPOINT")
	s3AccessKey := os.Getenv("CID_STORAGE_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("CID_STORAGE_S3_SECRET_KEY")

	if s3Endpoint != "" && s3AccessKey != "" && s3SecretKey != "" {
		client, err := storages3.NewS3Client(s3Endpoint, s3AccessKey, s3SecretKey, false)
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	return nil, nil
}
