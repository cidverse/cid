package storages3

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	minioClient *minio.Client
}

func (s3 S3Client) GetObject(ctx context.Context, bucketName string, objectName string) (*minio.Object, error) {
	object, err := s3.minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s3 S3Client) PutObject(ctx context.Context, bucketName string, objectName string, filePath string, contentType string) error {
	_, err := s3.minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

func NewS3Client(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*S3Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &S3Client{minioClient: minioClient}, nil
}
