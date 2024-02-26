package models

// ConfigurationSettings stores the actual configuration of the application
type ConfigurationSettings struct {
	Version            string             `json:"Version"`            // Version of the Settings
	ValidationSettings ValidationSettings `json:"ValidationSettings"` // Validation settings configured for this client
	ReportSettings     ReportSettings     `json:"ReportSettings"`     // Settings for the report module
}

// ReportSettings stores the information
type ReportSettings struct {
	ReportExecutionWindow int `json:"ReportExecutionWindow"` // Report execution window in minutes
	SendOnReportNumber    int `json:"SendOnReportNumber"`    // Indicates the number of reports to send on (ex. 10000000)
}
