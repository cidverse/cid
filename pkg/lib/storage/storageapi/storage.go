package storageapi

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type API interface {
	GetObject(ctx context.Context, bucketName, objectName string) (*minio.Object, error)
	PutObject(ctx context.Context, bucketName, objectName, filePath, contentType string) error
}
