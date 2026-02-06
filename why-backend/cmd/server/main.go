package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"why-backend/internal/api"
	"why-backend/internal/config"
	"why-backend/internal/storage"
	"why-backend/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitProvider(ctx, cfg.OTLPEndpoint)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "Failed to shutdown OpenTelemetry", "error", err)
		}
	}()

	// Initialize database
	db, err := storage.InitDB(ctx, cfg.PostgresURL)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize database", "error", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize MinIO
	minioClient, err := storage.InitMinIO(ctx, cfg.MinIO)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize MinIO", "error", err)
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	// Create router
	router := api.NewRouter(db, minioClient, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		slog.InfoContext(ctx, "Starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "Server failed", "error", err)
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.InfoContext(ctx, "Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Server forced to shutdown", "error", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	slog.InfoContext(ctx, "Server exited")
}
