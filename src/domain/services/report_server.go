package services

import "github.com/OpenBanking-Brasil/MQD_Client/domain/models"

// ReportServer is the Interface trhat exposes the methods to interact with report server
type ReportServer interface {
	SendReport(report models.Report) error                             // Send the report
	LoadAPIConfigurationFile(filePath string) ([]byte, error)          // Loads the configuration file specified in the path
	LoadConfigurationSettings() (*models.ConfigurationSettings, error) // Loads the configuration settings from the configuration file
}
