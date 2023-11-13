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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Structure to store the different system metrics
type Measurement struct {
	Timestamp time.Time // Time stamp of the metric
	Memory    uint64    // memmory value for this timestamp
	CPU       float64   // CPU value for this timestamp
}

var (
	requests                   metric.Float64Counter // Stores the number of requests the application has received
	endpoint_requests          metric.Float64Counter // Stores the number of requests by endpoint / server
	endpoint_validation_errors metric.Float64Counter // Stores the number of validation errors by endpoint / server
	mutex                      = sync.Mutex{}        // Mutex for thread-safe access
	requestsReceived           = 0                   // Stores the number of requests received
	badRequestsReceived        = 0                   // Stores the number of bad requests errors
	measurements               = []Measurement{}     // Create slices to store memory and CPU usage values
	responseTime               = []time.Duration{}   // Creates a slice to store the response time duration of requests
)

// startMemoryCalculator Starts the memmory calculation for observability
// @author AB
// @params
// @return
func startMemoryCalculator() {
	// Specify the duration for which you want to collect memory statistics in each interval
	collectionDuration := 1 * time.Minute // Change this as needed

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
// @author AB
// @params
// measurements: Lists of measurements to calculate the average
// @return
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

// collectCPUUsage collects the current CPU usage as a percentage.
// @author AB
// @params
// @return
// float64: Current value of CPU usage
func collectCPUUsage() float64 {
	// You would need to implement the code to collect CPU usage here.
	// This could involve using external tools or libraries depending on your platform.
	// Example: return someValueFromMonitoringTool()
	return 0.0 // Placeholder value, replace with actual implementation
}

// StartOpenTelemetry Initializes the counters and OpenTelemetry exporter for the service
// @author AB
// @params
// @return
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

	// This is the equivalent of prometheus.NewCounterVec
	endpoint_requests, err = meter.Float64Counter(
		"endpoint_requests",
		metric.WithDescription("Endpoint Requests by Server"),
		metric.WithUnit("requests"),
	)

	// This is the equivalent of prometheus.NewCounterVec
	endpoint_validation_errors, err = meter.Float64Counter(
		"endpoint_validation_errors",
		metric.WithDescription("Endpoint validation errors by Server"),
		metric.WithUnit("errors"),
	)

	if err != nil {
		log.Fatal(err)
	}

	requests.Add(ctx, 0)
}

// GetOpentelemetryHandler Returns the specified handler to export metrics
// @author AB
// @params
// @return
// http.Handler handler that supports metric export
func GetOpentelemetryHandler() http.Handler {
	return promhttp.Handler()
}

// RecordResponseDuration records thee response duration for a specific request.
// @author AB
// @params
// startTime: Initial start time for the request
// @return
func RecordResponseDuration(startTime time.Time) {
	mutex.Lock()
	responseTime = append(responseTime, time.Since(startTime))
	mutex.Unlock()
}

// IncreaseRequestsReceived increses the number of requests received metric
// @author AB
// @params
// @return
func IncreaseRequestsReceived() {
	mutex.Lock()
	requestsReceived++
	requests.Add(context.Background(), 1)
	mutex.Unlock()
}

// IncreaseBadRequestsReceived increses the number of bad requests received metric
// @author AB
// @params
// @return
func IncreaseBadRequestsReceived() {
	mutex.Lock()
	badRequestsReceived++
	mutex.Unlock()
}

// IncreaseValidationResult increses the number validation result for a specific server / endpoint, if the validation is false
// endpoint_validation_errors will also be increased
// @author AB
// @params
// serverid: IIdentifier of the server
// endpointName: Nmae of the endpoint
// valid: Validation result
// @return
func IncreaseValidationResult(serverId string, endpointName string, valid bool) {
	mutex.Lock()

	endpoint_requests.Add(context.Background(), 1, metric.WithAttributes(attribute.Key("server.name").String(serverId), attribute.Key("endpoint").String(endpointName)))
	if !valid {
		endpoint_validation_errors.Add(context.Background(), 1, metric.WithAttributes(attribute.Key("server.name").String(serverId), attribute.Key("endpoint").String(endpointName)))
	}

	mutex.Unlock()
}

// GetAndCleanRequestsReceived returns and cleans the lists of requests
// @author AB
// @params
// @return
// int: Number of requests recevied in the period of time
func GetAndCleanRequestsReceived() int {
	mutex.Lock()
	defer func() {
		requestsReceived = 0
		mutex.Unlock()
	}()

	return requestsReceived
}

// GetAndCleanBadRequestsReceived returns and cleans the lists of bad requests
// @author AB
// @params
// @return
// int: Number of bad requests recevied in the period of time
func GetAndCleanBadRequestsReceived() int {
	mutex.Lock()
	defer func() {
		badRequestsReceived = 0
		mutex.Unlock()
	}()

	return badRequestsReceived
}

// GetAndCleanAverageMemmory returns and cleans the average memmory used during the interval time
// @author AB
// @params
// @return
// string: Avg memmory used
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

// GetAndCleanResponseTime Returns and clenas the metric fot average response time
// @author AB
// @params
// @return
// string: Avg memmory used
func GetAndCleanResponseTime() string {
	mutex.Lock()
	avgTime := calculateAverageDuration(responseTime)
	responseTime = []time.Duration{}
	mutex.Unlock()
	return fmt.Sprint(avgTime)
}

// calculateAverageMemory calculates the average memory usage from a slice of measurements.
// @author AB
// @params
// @return
// string: Avg memmory used
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
