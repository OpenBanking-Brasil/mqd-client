package models

import "strings"

// ConfigurationSettings stores the actual configuration of the application
type ConfigurationSettings struct {
	Version            string             `json:"Version"`            // Version of the Settings
	ValidationSettings ValidationSettings `json:"ValidationSettings"` // Validation settings configured for this client
	ReportSettings     ReportSettings     `json:"ReportSettings"`     // Settings for the report module
	SecuritySettings   SecuritySettings
}

// ReportSettings stores the information
type ReportSettings struct {
	ReportExecutionWindow int `json:"ReportExecutionWindow"` // Report execution window in minutes
	SendOnReportNumber    int `json:"SendOnReportNumber"`    // Indicates the number of reports to send on (ex. 10000000)
}

// SecuritySettings Stores security settings information
type SecuritySettings struct {
	AttributesToMask []string
}

// HaveToMask indicates if a property valued must be masked or not
//
// Parameters:
//   - attributeName: Name of the attribute to verify
//
// Returns:
//   - ConfigurationManager: new created Local result manager
func (ss *SecuritySettings) HaveToMask(attributeName string) bool {
	for _, value := range ss.AttributesToMask {
		if strings.EqualFold(value, attributeName) {
			return true
		}
	}

	return false
}
