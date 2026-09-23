package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// NewTracer registra como global un TracerProvider que exporta las trazas por OTLP gRPC
//
// Param - ctx: Contexto de arranque del servicio, usado solo para crear el exportador
// Param - serviceName: Nombre con el que aparecen las trazas del servicio en Jaeger
// Param - endpoint: Dirección host:puerto del colector OTLP gRPC, sin esquema y sin TLS
// Returns - func/error: La función de apagado, que vacía los lotes pendientes y debe llamarse al salir, o el error si falla el exportador
func NewTracer(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
	// Crea el exportador sin conectar todavía: un colector caído no impide arrancar
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	// Falla si las opciones del exportador no son válidas
	if err != nil {
		return nil, fmt.Errorf("creating otlp exporter: %w", err)
	}

	// Identifica el servicio en cada span que se exporte
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName))

	// Agrupa los spans en lotes para no hacer una llamada de red por span
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Registra el provider global que usan otelgrpc y otelpgx
	otel.SetTracerProvider(provider)

	// Propaga el contexto de traza y el baggage entre servicios con las cabeceras W3C
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}
