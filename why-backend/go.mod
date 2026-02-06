module why-backend

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v7 v7.0.66
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.49.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
	go.opentelemetry.io/otel/exporters/prometheus v0.46.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/sdk/metric v1.24.0
	golang.org/x/crypto v0.22.0
)
