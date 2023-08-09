package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var Requests metric.Float64Counter

//
func StartOpenTelemetry() {
	ctx := context.Background()

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("Motor de Qualidade de dados"),
		semconv.ServiceVersionKey.String("v0.0.1"),
	)

	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	meterProvider := sdk.NewMeterProvider(
		sdk.WithResource(resources),
		sdk.WithReader(exporter),
	)

	meter := meterProvider.Meter(
		"API",
		metric.WithInstrumentationVersion("v0.0.0"),
	)

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()

	// This is the equivalent of prometheus.NewCounterVec
	Requests, err = meter.Float64Counter(
		"request_count",
		metric.WithDescription("Incoming request count"),
		metric.WithUnit("request"),
	)

	if err != nil {
		log.Fatal(err)
	}
	Requests.Add(ctx, 0)
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:8081/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
