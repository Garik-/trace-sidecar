package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func initMeterProvider(s service) (func(), error) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("exporter error: %w", err)
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(s.name),
			semconv.ServiceInstanceID(s.instanceID),
			semconv.ServiceVersion(s.version),
			semconv.ServiceNamespace(s.namespace),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("resource merge error: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(r),
	)

	otel.SetMeterProvider(provider)

	return func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Printf("error shutting down meter provider: %v", err)
		}
	}, nil
}
