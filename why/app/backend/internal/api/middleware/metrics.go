package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	httpServerDuration metric.Float64Histogram
)

// InitMetrics initializes OpenTelemetry metrics
func InitMetrics(ctx context.Context) error {
	meter := otel.Meter("why-backend")

	var err error
	httpServerDuration, err = meter.Float64Histogram(
		"http_server_duration_milliseconds",
		metric.WithDescription("Duration of HTTP requests in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	return nil
}

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics after request completes
		duration := float64(time.Since(start).Milliseconds())

		attrs := []attribute.KeyValue{
			attribute.String("service_name", "why-backend"),
			attribute.String("http_method", c.Request.Method),
			attribute.String("http_target", c.Request.URL.Path),
			attribute.Int("http_status_code", c.Writer.Status()),
		}

		httpServerDuration.Record(c.Request.Context(), duration, metric.WithAttributes(attrs...))
	}
}
