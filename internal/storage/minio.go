package storage

import (
	"context"
	"io"
	"time"

	"designreview/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Store wraps object storage. Images live here (ADR-0002); Postgres holds rows.
type Store struct {
	client *minio.Client
	bucket string
}

func NewMinio(cfg config.Config) (*Store, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioUser, cfg.MinioPass, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Store{client: client, bucket: cfg.MinioBucket}, nil
}

// Put streams an upload to object storage.
func (s *Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// PresignedURL mints a 1h GET URL for <img> tags (see docs/api.md upload flow).
func (s *Store) PresignedURL(ctx context.Context, key string) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Hour, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
