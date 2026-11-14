package storage

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cidverse/cid/pkg/lib/storage/storageapi"
	"github.com/cidverse/cid/pkg/lib/storage/storages3"
)

// DefaultBucketName is used when CID_STORAGE_S3_BUCKET is not set.
const DefaultBucketName = "cidverse-cid"

// GetStorageApi returns an S3 storage client if CID_STORAGE_S3_{ENDPOINT,ACCESS_KEY,SECRET_KEY} are set, otherwise nil.
// CID_STORAGE_S3_REGION optionally sets the bucket region.
// CID_STORAGE_S3_USE_SSL controls TLS (default true); set it to false for plain-HTTP endpoints.
func GetStorageApi() (storageapi.API, error) {
	s3Endpoint := os.Getenv("CID_STORAGE_S3_ENDPOINT")
	s3AccessKey := os.Getenv("CID_STORAGE_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("CID_STORAGE_S3_SECRET_KEY")
	s3Region := os.Getenv("CID_STORAGE_S3_REGION")
	s3UseSSL := true
	if v := os.Getenv("CID_STORAGE_S3_USE_SSL"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CID_STORAGE_S3_USE_SSL value %q: %w", v, err)
		}
		s3UseSSL = parsed
	}

	if s3Endpoint != "" && s3AccessKey != "" && s3SecretKey != "" {
		client, err := storages3.NewS3Client(s3Endpoint, s3AccessKey, s3SecretKey, s3Region, s3UseSSL)
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	return nil, nil
}

// GetBucketName returns the bucket from CID_STORAGE_S3_BUCKET, falling back to DefaultBucketName.
func GetBucketName() string {
	if bucket := os.Getenv("CID_STORAGE_S3_BUCKET"); bucket != "" {
		return bucket
	}

	return DefaultBucketName
}
