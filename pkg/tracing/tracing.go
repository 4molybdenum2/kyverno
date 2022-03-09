package tracing

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func NewExporter(ctx context.Context) {
	/* some tracing exporter like jaegar*/
}

func NewTracerProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("TestService"),
	)

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
	)
}
