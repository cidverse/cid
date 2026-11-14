package storages3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	minioClient *minio.Client
}

func (s3 S3Client) CheckBucket(ctx context.Context, bucketName string) error {
	exists, err := s3.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to access bucket %q: %w", bucketName, err)
	}
	if !exists {
		return fmt.Errorf("bucket %q does not exist", bucketName)
	}

	return nil
}

func (s3 S3Client) GetObject(ctx context.Context, bucketName string, objectName string) (*minio.Object, error) {
	object, err := s3.minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s3 S3Client) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, contentType string) error {
	_, err := s3.minioClient.PutObject(ctx, bucketName, objectName, reader, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

func (s3 S3Client) PutObjectFile(ctx context.Context, bucketName string, objectName string, filePath string, contentType string) error {
	_, err := s3.minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

// NewS3Client creates an S3 client. With useSSL and !tlsVerify, any server certificate is accepted (self-signed / private CA).
func NewS3Client(endpoint, accessKeyID, secretAccessKey, region string, useSSL bool, tlsVerify bool) (*S3Client, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
	}
	if useSSL && !tlsVerify {
		// start from minio's default transport (timeouts, proxy, SSL_CERT_FILE) and only disable verification
		transport, err := minio.DefaultTransport(true)
		if err != nil {
			return nil, fmt.Errorf("failed to create s3 transport: %w", err)
		}
		transport.TLSClientConfig.InsecureSkipVerify = true // #nosec G402 -- explicit opt-in via CID_STORAGE_S3_TLS_VERIFY=false
		opts.Transport = transport
	}

	minioClient, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}

	return &S3Client{minioClient: minioClient}, nil
}
