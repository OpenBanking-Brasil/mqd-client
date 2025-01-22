package models

import "time"

// MetricObject Contains the name and value for different types of metrics for the report
type MetricObject struct {
	Key   string // Name of the metric
	Value string // Value of the metric
}

// ApplicationMetrics Contains a list of metrics recorded for the report
type ApplicationMetrics struct {
	Values []MetricObject // List of metrics with its values
}

// ConfigurationUpdateError Stores the information for the configuration update errors
type ConfigurationUpdateError struct {
	ApplicationConfigurationID uint
	ErrorDate                  time.Time
	ErrorMessage               string
}

// ConfigurationUpdateStatus Stores the information for the configuration update status
type ConfigurationUpdateStatus struct {
	ConfigurationVersion     string                     // Version of the configuration
	LastExecutionDate        time.Time                  // Indicates the data execution of the configuration update
	LastUpdatedDate          time.Time                  // Indicates the data of the las successful configuration update
	ConfigurationUpdateError []ConfigurationUpdateError // List of error messages if any durin the update process
}

// ApplicationConfiguration Contains the information of the actual configuration of the application
type ApplicationConfiguration struct {
	ApplicationVersion               string                    // Version of the application
	Environment                      string                    // Environment of the application
	ConfigurationUpdateStatus        ConfigurationUpdateStatus // Status of the configuration update
	ReportExecutionWindow            string                    // Report Execution Window of the application
	ReportExecutionNumber            string                    // Report Execution Number limit of the application
	ApplicationMode                  string                    // Mode of the application - TRANSMITTER / RECEIVER
	ApplicationID                    string                    // unique identifier for the application
	ResultSettingsEnabled            bool                      // ResultSettings Enabled
	ResultSettingsFilesPerDay        int                       // ResultSettings Files per day
	ResultSettingsDaysToStore        int                       // ResultSettings Days To Store
	ResultSettingsSamplesPerError    int                       // ResultSettings Samples Per Error
	ResultSettingsMaskPrivateContent bool                      // ResultSettings Mask PrivateContent
}

// UnsupportedEndpoint shows the list of unsupported endpoints requested to the API
type UnsupportedEndpoint struct {
	EndpointName string `json:"EndpointName"` // EndpointName shows the name of the endpoint requested
	Version      string `json:"Version"`      // Version shows the version of the endpoint requested
	Count        int    `json:"Count"`        // Count shows the number of times the endpoint was requested
	Error        string `json:"Error"`        // Error shows the error message if any
}

// ServerSummary contains Summary of a specific server
type ServerSummary struct {
	ServerID        string            // Server identifier (UUID)
	TotalRequests   int               // Total number of requests
	EndpointSummary []EndPointSummary // Summary of the endpoints requested
}

// FieldDetail contains the details for a filed with an error type
type FieldDetail struct {
	ErrorType  string   // Name of the error type found
	XFapiList  []string // List of xFapiInteractionIds that showed this specific error
	TotalCount int      // Number of times the error was found
}

// EndPointSummaryDetail contains the Summary of the details of errors for a specific field
type EndPointSummaryDetail struct {
	Field   string        // Name of the field
	Details []FieldDetail // List of details with the errors
}

// EndPointSummary Contains a summary for a specific endpoint
type EndPointSummary struct {
	EndpointName     string                  // Name of the endpoint
	TotalRequests    int                     // Total number of requests
	ValidationErrors int                     // Total number of validation errors
	Detail           []EndPointSummaryDetail // Detail of the errors
}

// Report is the object to be sent to the server
type Report struct {
	Metrics                  ApplicationMetrics       // Metrics of the application
	ApplicationConfiguration ApplicationConfiguration // Configuration of the application on the Client Side
	ClientID                 string                   // Client identifier (UUID)
	DataOwnerID              string                   // OrganisationID of the institution reporting the information
	UnsupportedEndpoints     []UnsupportedEndpoint    // List with the unsupported endpoint requests
	ServerSummary            []ServerSummary          // List of Servers requested
}
