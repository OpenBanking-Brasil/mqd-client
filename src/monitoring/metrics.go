package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type Measurement struct {
	Timestamp time.Time
	Memory    uint64
	CPU       float64
}

var (
	requests            metric.Float64Counter
	mutex               = sync.Mutex{} // Mutex for thread-safe access
	requestsReceived    = 0
	badRequestsReceived = 0
	measurements        = []Measurement{} // Create slices to store memory and CPU usage values
	responseTime        = []time.Duration{}
)

func startMemoryCalculator() {
	// Specify the duration for which you want to collect memory statistics in each interval
	collectionDuration := 10 * time.Second // Change this as needed

	// Create a ticker to trigger data collection at the specified interval
	ticker := time.NewTicker(collectionDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			// Collect memory and CPU statistics for the specified duration
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			cpuUsage := collectCPUUsage()

			// Append the measurements to the slice
			measurements = append(measurements, Measurement{
				Timestamp: time.Now(),
				Memory:    memStats.Alloc,
				CPU:       cpuUsage,
			})
			mutex.Unlock()
		}
	}
}

// calculateAverageMemory calculates the average memory usage from a slice of measurements.
func calculateAverageMemory(measurements []Measurement) uint64 {
	if len(measurements) == 0 {
		return 0
	}
	var sum uint64
	for _, m := range measurements {
		sum += m.Memory
	}
	return sum / uint64(len(measurements))
}

// calculateAverageCPU calculates the average CPU usage from a slice of measurements.
func calculateAverageCPU(measurements []Measurement) float64 {
	if len(measurements) == 0 {
		return 0
	}
	var sum float64
	for _, m := range measurements {
		sum += m.CPU
	}
	return sum / float64(len(measurements))
}

// collectCPUUsage collects the current CPU usage as a percentage.
func collectCPUUsage() float64 {
	// You would need to implement the code to collect CPU usage here.
	// This could involve using external tools or libraries depending on your platform.
	// Example: return someValueFromMonitoringTool()
	return 0.0 // Placeholder value, replace with actual implementation
}

func StartOpenTelemetry() {
	ctx := context.Background()
	go startMemoryCalculator()

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

	// This is the equivalent of prometheus.NewCounterVec
	requests, err = meter.Float64Counter(
		"request_count",
		metric.WithDescription("Incoming request count"),
		metric.WithUnit("request"),
	)

	if err != nil {
		log.Fatal(err)
	}
	requests.Add(ctx, 0)
}

func GetOpentelemetryHandler() http.Handler {
	return promhttp.Handler()
	//http.Handle("/metrics", promhttp.Handler())
}

func RecordResponseDuration(startTime time.Time) {
	mutex.Lock()
	responseTime = append(responseTime, time.Since(startTime))
	mutex.Unlock()
}

func IncreaseRequestsReceived() {
	mutex.Lock()
	requestsReceived++
	requests.Add(context.Background(), 1)
	mutex.Unlock()
}

func IncreaseBadRequestsReceived() {
	mutex.Lock()
	badRequestsReceived++
	mutex.Unlock()
}

func GetAndCleanRequestsReceived() int {
	mutex.Lock()
	defer func() {
		requestsReceived = 0
		mutex.Unlock()
	}()

	return requestsReceived
}

func GetAndCleanBadRequestsReceived() int {
	mutex.Lock()
	defer func() {
		badRequestsReceived = 0
		mutex.Unlock()
	}()

	return badRequestsReceived
}

func GetAndCleanAverageMemmory() string {
	mutex.Lock()
	// Calculate the average memory usage and CPU consumption and print them
	avgMemory := calculateAverageMemory(measurements)
	// Reset measurements for the next interval
	measurements = []Measurement{}
	mutex.Unlock()
	result := fmt.Sprintf("%.2f MB", float64(avgMemory)/1024/1024)
	return result
}

func GetAndCleanResponseTime() string {
	mutex.Lock()
	avgTime := calculateAverageDuration(responseTime)
	responseTime = []time.Duration{}
	mutex.Unlock()
	return fmt.Sprint(avgTime)
}

// calculateAverageMemory calculates the average memory usage from a slice of measurements.
func calculateAverageDuration(durations []time.Duration) int64 {
	if len(durations) == 0 {
		return 0
	}
	var sum int64
	for _, m := range durations {
		sum += m.Microseconds()
	}
	return sum / int64(len(durations))
}
