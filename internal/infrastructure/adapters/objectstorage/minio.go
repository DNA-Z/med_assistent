// Package objectstorage содержит S3-совместимый адаптер хранения файлов.
package objectstorage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// MinIO реализует ObjectStorage поверх MinIO или другого S3-совместимого сервиса.
type MinIO struct {
	client *minio.Client
	bucket string
}

// NewMinIO создаёт клиент и при необходимости создаёт бакет.
func NewMinIO(ctx context.Context, cfg Config) (*MinIO, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), Secure: cfg.UseSSL})
	if err != nil {
		return nil, fmt.Errorf("создать S3-клиент: %w", err)
	}
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("проверить бакет %q: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("создать бакет %q: %w", cfg.Bucket, err)
		}
	}
	return &MinIO{client: client, bucket: cfg.Bucket}, nil
}

func (s *MinIO) Put(ctx context.Context, object ports.StoredObject) error {
	if object.ContentType == "" {
		object.ContentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, object.Key, object.Reader, object.Size, minio.PutObjectOptions{ContentType: object.ContentType})
	if err != nil {
		return fmt.Errorf("сохранить объект %q: %w", object.Key, err)
	}
	return nil
}

func (s *MinIO) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("открыть объект %q: %w", key, err)
	}
	if _, err := object.Stat(); err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("проверить объект %q: %w", key, err)
	}
	return object, nil
}

func (s *MinIO) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("удалить объект %q: %w", key, err)
	}
	return nil
}

var _ ports.ObjectStorage = (*MinIO)(nil)
