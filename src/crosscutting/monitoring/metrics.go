package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
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

// Version indicates the current version of the application
const Version = "2.3.1"

// Measurement is a Structure to store the different system metrics
type Measurement struct {
	Timestamp     time.Time // Time stamp of the metric
	Memory        uint64    // memory value for this timestamp
	MaxUSedMemory uint64    // max memory value for this timestamp
	CPU           float64   // CPU value for this timestamp
	NumCPU        int
}

// SystemMetrics information about system metrics
type SystemMetrics struct {
	AverageMemory       string
	MaxUsedMemory       string
	CPUUsage            string
	AllowedCPUs         string
	RequestsReceived    string
	BadRequestsReceived string
	AverageResponseTime string
}

var (
	requests                 metric.Float64Counter // Stores the number of requests the application has received
	endpointRequests         metric.Float64Counter // Stores the number of requests by endpoint / server
	endpointValidationErrors metric.Float64Counter // Stores the number of validation errors by endpoint / server
	mutex                    = sync.Mutex{}        // Mutex for thread-safe access
	requestsReceived         = 0                   // Stores the number of requests received
	badRequestsReceived      = 0                   // Stores the number of bad requests errors
	measurements             []Measurement
	responseTime             []time.Duration
	unsupportedEndpoints     = make(map[string]map[string]int) // Stores the number of unsupported endpoints
)

// startMemoryCalculator Starts the memory calculation for observability
//
// Parameters:
//
// Returns:
func startMemoryCalculator() {
	// Specify the duration for which you want to collect memory statistics in each interval
	collectionDuration := 1 * time.Minute // Change this as needed

	// Create a ticker to trigger data collection at the specified interval
	ticker := time.NewTicker(collectionDuration)
	defer ticker.Stop()

	for range ticker.C {
		mutex.Lock()
		// Collect memory and CPU statistics for the specified duration
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		cpuUsage := collectCPUUsage()

		// Append the measurements to the slice
		measurements = append(measurements, Measurement{
			Timestamp:     time.Now(),
			Memory:        memStats.Alloc,
			MaxUSedMemory: memStats.TotalAlloc,
			CPU:           cpuUsage,
			NumCPU:        runtime.NumCPU(),
		})
		mutex.Unlock()
	}
}

// calculateAverageMemory calculates the average memory usage from a slice of measurements.
//
// Parameters:
//   - measurements: Lists of measurements to calculate the average
//
// Returns:
//   - uint64: Average memory used
func calculateAverageMemory(measurements []Measurement) (uint64, uint64, int) {
	if len(measurements) == 0 {
		return 0, 0, 0
	}
	var sum uint64
	var maxMemory uint64
	maxMemory = 0
	maxCPU := 0
	for _, m := range measurements {
		sum += m.Memory
		if m.MaxUSedMemory > maxMemory {
			maxMemory = m.MaxUSedMemory
		}

		if m.NumCPU >= maxCPU {
			maxCPU = m.NumCPU
		}
	}

	return sum / uint64(len(measurements)), maxMemory, maxCPU
}

// collectCPUUsage collects the current CPU usage as a percentage.
//
// Parameters:
//
// Returns:
//   - float64: Average CPU used
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
		semconv.ServiceVersionKey.String(Version),
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

	// This is the equivalent of prometheus.NewCounterVec
	endpointRequests, err = meter.Float64Counter(
		"endpoint_requests",
		metric.WithDescription("Endpoint Requests by Server"),
		metric.WithUnit("requests"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// This is the equivalent of prometheus.NewCounterVec
	endpointValidationErrors, err = meter.Float64Counter(
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

// IncreaseRequestsReceived increases the number of requests received metric
// @author AB
// @params
// @return
func IncreaseRequestsReceived() {
	mutex.Lock()
	requestsReceived++
	requests.Add(context.Background(), 1)
	mutex.Unlock()
}

// IncreaseBadRequestsReceived increases the number of bad requests received metric
// @author AB
// @params
// @return
func IncreaseBadRequestsReceived() {
	mutex.Lock()
	badRequestsReceived++
	mutex.Unlock()
}

// IncreaseBadEndpointsReceived increases the number of bad requests received metric
//
// Parameters:
//   - endpoint: Endpoint name
//   - errorMessage: API Version of the endpoint
//   - float64: Error message to record
//
// Returns:
func IncreaseBadEndpointsReceived(endpoint string, version string, errorMessage string) {
	mutex.Lock()
	badRequestsReceived++
	if unsupportedEndpoints[endpoint] == nil {
		unsupportedEndpoints[endpoint] = make(map[string]int)
	}

	unsupportedEndpoints[endpoint][version]++
	mutex.Unlock()
}

// IncreaseValidationResult increases the number validation result for a specific server / endpoint, if the validation is false
// endpoint_validation_errors will also be increased
//
// Parameters:
//   - serverID: Identifier of the server
//   - endpointName: Name of the endpoint
//   - valid: Validation result
//
// Returns:
func IncreaseValidationResult(serverID string, endpointName string, valid bool) {
	mutex.Lock()

	endpointRequests.Add(context.Background(), 1, metric.WithAttributes(attribute.Key("server.name").String(serverID), attribute.Key("endpoint").String(endpointName)))
	if !valid {
		endpointValidationErrors.Add(context.Background(), 1, metric.WithAttributes(attribute.Key("server.name").String(serverID), attribute.Key("endpoint").String(endpointName)))
	}

	mutex.Unlock()
}

// GetAndCleanRequestsReceived returns and cleans the lists of requests
// @author AB
// @params
// @return
// int: Number of requests received in the period of time
func getAndCleanRequestsReceived() int {
	defer func() {
		requestsReceived = 0
	}()

	return requestsReceived
}

// GetAndCleanBadRequestsReceived returns and cleans the lists of bad requests
// @author AB
// @params
// @return
// int: Number of bad requests received in the period of time
func getAndCleanBadRequestsReceived() int {
	defer func() {
		badRequestsReceived = 0
	}()

	return badRequestsReceived
}

// GetAndCleanUnsupportedEndpoints returns and cleans the lists of bad requests
// endpoint_validation_errors will also be increased
//
// Parameters:
//
// Returns:
//   - map: map[string]map[string]int Number of bad requests received in the period of time by endpoint and version
func GetAndCleanUnsupportedEndpoints() map[string]map[string]int {
	mutex.Lock()
	defer func() {
		unsupportedEndpoints = make(map[string]map[string]int)
		mutex.Unlock()
	}()

	return unsupportedEndpoints
}

// GetAndCleanResponseTime Returns and cleans the metric fot average response time
// @author AB
// @params
// @return
// string: Avg memory used
func getAndCleanResponseTime() string {
	avgTime := calculateAverageDuration(responseTime)
	responseTime = []time.Duration{}
	return fmt.Sprint(avgTime)
}

// calculateAverageMemory calculates the average memory usage from a slice of measurements.
// @author AB
// @params
// @return
// string: Avg memory used
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

// GetAndCleanSystemMetrics returns and cleans all system metrics
// @author AB
// @params
// @return
// SystemMetrics: instance of the system metrics object
func GetAndCleanSystemMetrics() SystemMetrics {
	mutex.Lock()
	// Calculate the average memory usage and CPU consumption and print them
	avgMemory, maxMemory, numCPU := calculateAverageMemory(measurements)

	result := SystemMetrics{
		AverageMemory:       fmt.Sprintf("%.2f MB", float64(avgMemory)/1024/1024),
		MaxUsedMemory:       fmt.Sprintf("%.2f MB", float64(maxMemory)/1024/1024),
		CPUUsage:            "",
		AllowedCPUs:         strconv.Itoa(numCPU),
		RequestsReceived:    strconv.Itoa(getAndCleanRequestsReceived()),
		BadRequestsReceived: strconv.Itoa(getAndCleanBadRequestsReceived()),
		AverageResponseTime: getAndCleanResponseTime(),
	}

	// Reset measurements for the next interval
	measurements = []Measurement{}
	mutex.Unlock()

	return result
}
