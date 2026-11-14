package storageapi

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type API interface {
	// CheckBucket verifies that the storage is reachable, the credentials are valid and the bucket exists.
	CheckBucket(ctx context.Context, bucketName string) error
	GetObject(ctx context.Context, bucketName, objectName string) (*minio.Object, error)
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, contentType string) error
	PutObjectFile(ctx context.Context, bucketName, objectName, filePath, contentType string) error
}
