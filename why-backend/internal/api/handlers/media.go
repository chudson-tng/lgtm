package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"why-backend/internal/config"
	"why-backend/internal/storage"
)

var mediaTracer = otel.Tracer("why-backend/handlers/media")

type MediaHandler struct {
	minio  *minio.Client
	config *config.Config
}

func NewMediaHandler(minio *minio.Client, cfg *config.Config) *MediaHandler {
	return &MediaHandler{
		minio:  minio,
		config: cfg,
	}
}

// UploadMedia handles file uploads to MinIO
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	ctx, span := mediaTracer.Start(c.Request.Context(), "UploadMedia")
	defer span.End()

	file, err := c.FormFile("file")
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	span.SetAttributes(
		attribute.String("file.name", file.Filename),
		attribute.Int64("file.size", file.Size),
	)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to open uploaded file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	contentType := storage.GetContentType(file.Filename)

	// Upload to MinIO
	url, err := storage.UploadFile(ctx, h.minio, h.config.MinIO.BucketName, objectName, src, file.Size, contentType)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to upload file to MinIO", "error", err, "filename", file.Filename)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	span.SetAttributes(attribute.String("object.url", url))
	slog.InfoContext(ctx, "File uploaded successfully", "url", url, "size", file.Size)

	c.JSON(http.StatusOK, gin.H{"url": url})
}
