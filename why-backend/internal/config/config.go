package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	PostgresURL  string
	MinIO        MinIOConfig
	OTLPEndpoint string
	JWTSecret    string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		PostgresURL:  getEnv("POSTGRES_URL", ""),
		OTLPEndpoint: getEnv("OTLP_ENDPOINT", "alloy.observability.svc.cluster.local:4317"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "loki-minio.observability.svc.cluster.local:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "loki"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "supersecret"),
			BucketName:      getEnv("MINIO_BUCKET", "why-media"),
			UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
		},
	}

	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
