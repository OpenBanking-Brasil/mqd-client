package configuration

import (
	"context"
	"fmt"
	"os"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/google/uuid"
	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

const (
	serverOrgIDEnv     = "SERVER_ORG_ID"    // constant  to store name of the server id environment variable
	applicationModeEnv = "APPLICATION_MODE" // constant  to store name of the application mode environment variable"
	transmitterMode    = "TRANSMITTER"      // TRANSMITTER Application mode Constant
	receiverMode       = "RECEIVER"         // RECEIVER Application mode Constant
	certPath           = "/certificates/"
)

var (
	// ServerID has the OrganisationID for the server
	ServerID = ""
)

// Configuration exposes the settings of the application
type Configuration struct {
	logger   log.Logger
	Settings Settings
}

// GetApplicationSettings Loads all settings required for the application to run, such as endpoint settings and environment settings
//
// Parameters:
// Returns:
func (cnf *Configuration) GetApplicationSettings() Settings {
	cnf.logger = log.GetLogger()
	cnf.logger.Info("Initializing application configuration", "configuration", "GetApplicationSettings")
	err := cnf.loadApplicationSettings()
	if err != nil {
		cnf.logger.Fatal(err, "Error initializing application configuration", "configuration", "GetApplicationSettings")
	}

	if !cnf.validateSettings() {
		cnf.logger.Fatal(err, "Please correct the problems with the validation settings", "configuration", "GetApplicationSettings")
	}

	cnf.Settings.ConfigurationSettings.ApplicationID = uuid.New()
	return cnf.Settings
}

// loadApplicationSettings Loads all settings required for the application to run, such as endpoint settings and environment settings
//
// Parameters:
// Returns:
func (cnf *Configuration) loadApplicationSettings() error {
	err := cnf.loadConfigurationFile()
	if err != nil {
		cnf.logger.Error(err, "Error initializing application configuration, using only environment values", "configuration", "loadApplicationSettings")
	}

	err = cnf.loadSettingsFromEnvironment()
	if err != nil {
		return err
	}

	return nil
}

// validateSettings Validates the loaded settings with the allowed values
//
// Parameters:
// Returns: true if validation was ok
func (cnf *Configuration) validateSettings() bool {
	isValid := true
	if !(cnf.Settings.ApplicationSettings.Mode == transmitterMode || cnf.Settings.ApplicationSettings.Mode == receiverMode) {
		cnf.logger.Warning("APPLICATION_MODE not found, please set Environment Variable: ["+applicationModeEnv+"], as ["+transmitterMode+"] or ["+receiverMode+"] ", "Configuration", "validateSettings")
		isValid = false
	}

	_, err := uuid.Parse(cnf.Settings.ApplicationSettings.OrganisationID)
	if err != nil {
		cnf.logger.Warning("ClientID not found or wrong format, please set Environment Variable: ["+serverOrgIDEnv+"], or OrganisationID variable on configuration file", "Configuration", "validateSettings")
		isValid = false
	}

	if cnf.Settings.ReportSettings.ExecutionWindow != 0 && (cnf.Settings.ReportSettings.ExecutionWindow > 60 || cnf.Settings.ReportSettings.ExecutionWindow < 0) {
		cnf.logger.Warning("Value out of range for  REPORT_EXECUTION_WINDOW(1 - 60), using default value from system", "Configuration", "validateSettings")
		cnf.Settings.ReportSettings.ExecutionWindow = 0
	}

	if cnf.Settings.ReportSettings.ExecutionNumber != 0 && (cnf.Settings.ReportSettings.ExecutionNumber > 200000 || cnf.Settings.ReportSettings.ExecutionNumber < 10000) {
		cnf.logger.Warning("Value out of range for REPORT_EXECUTION_NUMBER (10000 - 200000), using default value from system", "Configuration", "validateSettings")
		cnf.Settings.ReportSettings.ExecutionNumber = 0
	}

	if cnf.Settings.SecuritySettings.EnableHTTPS {
		cnf.validateHTTPSCertificates()
	}

	if cnf.Settings.ResultSettings.FilesPerDay < 1 || cnf.Settings.ResultSettings.FilesPerDay > 24 {
		cnf.logger.Warning("Value out of range for RESULT_FILES_PER_DAY (1 - 24), using default value from system", "Configuration", "validateSettings")
		cnf.Settings.ResultSettings.FilesPerDay = 8
	}

	if cnf.Settings.ResultSettings.SamplesPerError < 1 || cnf.Settings.ResultSettings.SamplesPerError > 10 {
		cnf.logger.Warning("Value out of range for RESULT_SAMPLES_PER_ERROR (1 - 10), using default value from system", "Configuration", "validateSettings")
		cnf.Settings.ResultSettings.SamplesPerError = 5
	}

	if cnf.Settings.ResultSettings.DaysToStore < 1 || cnf.Settings.ResultSettings.DaysToStore > 10 {
		cnf.logger.Warning("Value out of range for RESULT_DAYS_TO_STORE (1 - 10), using default value from system", "Configuration", "validateSettings")
		cnf.Settings.ResultSettings.SamplesPerError = 7
	}

	return isValid
}

func (cnf *Configuration) validateHTTPSCertificates() bool {
	certFile := "server.crt"
	keyFile := "server.key"

	cnf.Settings.SecuritySettings.KeyFilePath = fmt.Sprintf("%s%s", certPath, keyFile)
	cnf.Settings.SecuritySettings.CertFilePath = fmt.Sprintf("%s%s", certPath, certFile)
	_, err := os.Stat(cnf.Settings.SecuritySettings.KeyFilePath)
	if os.IsNotExist(err) {
		cnf.logger.Panic("Key certificate not found: "+cnf.Settings.SecuritySettings.KeyFilePath, "Configuration", "validateHTTPSCertificates")
	}

	_, err = os.Stat(cnf.Settings.SecuritySettings.CertFilePath)
	if os.IsNotExist(err) {
		cnf.logger.Panic("Certificate file not found: "+cnf.Settings.SecuritySettings.CertFilePath, "Configuration", "validateHTTPSCertificates")
	}

	return true
}

// loadConfigurationFile Loads the settings from the configuration file
//
// Parameters:
// Returns: Error if any
func (cnf *Configuration) loadConfigurationFile() error {
	cnf.logger.Info("Loading configuration file", "configuration", "loadConfigurationFile")
	fileName := "./settings/settings.yml"

	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		cnf.logger.Warning("Settings file not found. using only environment values", "configuration", "loadConfigurationFile")
		return nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		cnf.logger.Error(err, "There was an error loading the configuration File.", "configuration", "loadConfigurationFile")
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cnf.Settings)
	if err != nil {
		cnf.logger.Error(err, "There was an error while reading the configuration File.", "configuration", "loadConfigurationFile")
		return err
	}

	return nil
}

// loadSettingsFromEnvironment Loads the application settings from the environment and overwrites the settings from the file
//
// Parameters:
// Returns: Error if any
func (cnf *Configuration) loadSettingsFromEnvironment() error {
	cnf.logger.Info("Loading configuration from environment", "configuration", "loadSettingsFromEnvironment")
	ctx := context.Background()
	err := envconfig.Process(ctx, &cnf.Settings)
	if err != nil {
		cnf.logger.Error(err, "There was an error processing environment settings.", "configuration", "loadSettingsFromEnvironment")
	}

	return nil
}
