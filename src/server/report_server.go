package server

// Contains the name and value for different types of metrics for the report
type MetricObject struct {
	Key   string // Name of the metric
	Value string // Value of the metric
}

// Contains a list of metrics recorded for the report
type ApplicationMetrics struct {
	Values []MetricObject // List of metrics with its values
}

// Detail for a filed with an error type
type FieldDetail struct {
	ErrorType  string   // Name of the error type found
	XFapiList  []string // List of xFapiInteractionIds that showed this specific error
	TotalCount int      // Number of times the error was found
}

// Summary of the details of errors for a specific field
type EndPointSummaryDetail struct {
	Field   string        // Name of the field
	Details []FieldDetail // List of details with the errors
}

// Contains a summary for a specific endpoint
type EndPointSummary struct {
	EndpointName     string                  // Name of the endpoint
	TotalRequests    int                     // Totla number of requests
	ValidationErrors int                     // Total number of validation errors
	Detail           []EndPointSummaryDetail // Detail of the errors
}

// Summary of a specific server
type ServerSummary struct {
	ServerId        string            // Server identifier (UUID)
	TotalRequests   int               // Total number of requests
	EndpointSummary []EndPointSummary // Summary of the endpoints requested
}

// Object report to be sent to the server
type Report struct {
	Metrics       ApplicationMetrics // Metris of the application
	ClientID      string             // Client identifier (UUID)
	ServerSummary []ServerSummary    // List of Servers requested
}

// Interface for the report server
type ReportServer interface {
	SendReport(report Report) error // Send the report
	LoadAPIConfigurationFile(filePath string) ([]byte, error)
}
