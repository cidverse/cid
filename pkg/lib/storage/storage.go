package storage

import (
	"fmt"
	"log/slog"
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
// CID_STORAGE_S3_TLS_VERIFY controls certificate verification (default true); set it to false to accept any certificate.
func GetStorageApi() (storageapi.API, error) {
	s3Endpoint := os.Getenv("CID_STORAGE_S3_ENDPOINT")
	s3AccessKey := os.Getenv("CID_STORAGE_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("CID_STORAGE_S3_SECRET_KEY")
	s3Region := os.Getenv("CID_STORAGE_S3_REGION")
	s3UseSSL, err := envBool("CID_STORAGE_S3_USE_SSL", true)
	if err != nil {
		return nil, err
	}
	s3TLSVerify, err := envBool("CID_STORAGE_S3_TLS_VERIFY", true)
	if err != nil {
		return nil, err
	}

	if s3Endpoint != "" && s3AccessKey != "" && s3SecretKey != "" {
		if s3UseSSL && !s3TLSVerify {
			slog.Warn("S3 TLS certificate verification is disabled (CID_STORAGE_S3_TLS_VERIFY=false)", "endpoint", s3Endpoint)
		}
		client, err := storages3.NewS3Client(s3Endpoint, s3AccessKey, s3SecretKey, s3Region, s3UseSSL, s3TLSVerify)
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

// envBool parses a boolean env variable, returning def if it is unset.
func envBool(name string, def bool) (bool, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid %s value %q: %w", name, v, err)
	}

	return parsed, nil
}
