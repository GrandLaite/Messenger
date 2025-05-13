package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	client *minio.Client
	logger *slog.Logger
}

func NewStorageService(endpoint, accessKey, secretKey string, useSSL bool, logger *slog.Logger) (*StorageService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &StorageService{
		client: client,
		logger: logger,
	}, nil
}

func (s *StorageService) CreateBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		s.logger.Info("Bucket created", "bucket", bucketName)
	}
	return nil
}

func (s *StorageService) UploadFile(ctx context.Context, bucketName string, file io.Reader, fileSize int64, contentType string) (string, error) {
	objectName := uuid.New().String()

	info, err := s.client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	s.logger.Info("File uploaded", "object", info.Key)
	return objectName, nil
}

func (s *StorageService) DownloadFile(ctx context.Context, bucketName, objectName string) (*minio.Object, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	_, err = obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, nil
}
