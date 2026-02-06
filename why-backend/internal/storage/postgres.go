package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("why-backend/storage")

// InitDB initializes the PostgreSQL connection and runs migrations
func InitDB(ctx context.Context, postgresURL string) (*sql.DB, error) {
	ctx, span := tracer.Start(ctx, "InitDB")
	defer span.End()

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(ctx, db); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	span.SetAttributes(attribute.Bool("migrations.success", true))
	return db, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	ctx, span := tracer.Start(ctx, "runMigrations")
	defer span.End()

	// Read migration file
	migrationPath := filepath.Join("migrations", "001_create_schema.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := db.ExecContext(ctx, string(migrationSQL)); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	span.SetAttributes(attribute.String("migration.file", migrationPath))
	return nil
}
