package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitProvider initializes OpenTelemetry with traces, metrics, and logs
func InitProvider(ctx context.Context, otlpEndpoint string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	}

	// Create resource with service information (best practice)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("why-backend"),
			semconv.ServiceVersion("1.0.0"),
			semconv.ServiceInstanceID(os.Getenv("HOSTNAME")), // Pod name
			semconv.DeploymentEnvironment("production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// ========== TRACES ==========
	// OTLP trace exporter to Alloy
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces for demo
	)

	otel.SetTracerProvider(tracerProvider)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	// ========== METRICS ==========
	// Prometheus exporter for metrics scraping by Alloy
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(promExporter),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	// ========== CONTEXT PROPAGATION ==========
	// W3C Trace Context propagation (best practice)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// ========== STRUCTURED LOGGING ==========
	// Configure structured JSON logging to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.InfoContext(ctx, "OpenTelemetry initialized",
		"service", "why-backend",
		"version", "1.0.0",
		"otlp_endpoint", otlpEndpoint,
	)

	return shutdown, nil
}
