package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
	"why-backend/internal/config"
)

// InitMinIO initializes the MinIO client and ensures the bucket exists
func InitMinIO(ctx context.Context, cfg config.MinIOConfig) (*minio.Client, error) {
	ctx, span := tracer.Start(ctx, "InitMinIO")
	defer span.End()

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Create bucket if it doesn't exist
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		span.SetAttributes(attribute.Bool("bucket.created", true))
	}

	span.SetAttributes(
		attribute.String("bucket.name", cfg.BucketName),
		attribute.String("endpoint", cfg.Endpoint),
	)

	return client, nil
}

// UploadFile uploads a file to MinIO and returns its URL
func UploadFile(ctx context.Context, client *minio.Client, bucketName, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	ctx, span := tracer.Start(ctx, "UploadFile")
	defer span.End()

	span.SetAttributes(
		attribute.String("object.name", objectName),
		attribute.Int64("object.size", size),
		attribute.String("content.type", contentType),
	)

	_, err := client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the URL to access the file
	url := fmt.Sprintf("http://%s/%s/%s", client.EndpointURL().Host, bucketName, objectName)
	span.SetAttributes(attribute.String("object.url", url))

	return url, nil
}

// GetContentType returns the MIME type based on file extension
func GetContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	default:
		return "application/octet-stream"
	}
}
