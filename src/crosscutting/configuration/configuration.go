package configuration

import (
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/google/uuid"
)

const (
	serverOrgIDEnv     = "SERVER_ORG_ID"    // constant  to store name of the server id environment variable
	loggingLevelEnv    = "LOGGING_LEVEL"    // constant  to store name of the Logging level environment variable
	environmentEnv     = "ENVIRONMENT"      // constant  to store name of the environment variable
	applicationModeEnv = "APPLICATION_MODE" // constant  to store name of the application mode environment variable"
	transmitterMode    = "TRANSMITTER"      // TRANSMITTER Application mode Constant
	receiverMode       = "RECEIVER"         // RECEIVER Application mode Constant
	proxyURL           = "PROXY_URL"        // RECEIVER Application mode Constant
)

var (
	// ServerID has the OrganisationID for the server
	ServerID = ""
	// ClientID id for the client
	ClientID = ""
	// Environment Indicates the actual Environment the app is running
	Environment = ""
	// ApplicationMode Indicates the actual Application Mode the app is running in (TRANSMISSOR, RECEPTOR)
	ApplicationMode         = ""
	lastReportExecutionDate = time.Time{} // Indicates the data of the last report execution
	// ServerURL to send the reports to
	ServerURL       = ""
	lastUpdatedDate = time.Time{} // Indicates the data of the last report update
	ApplicationID   = uuid.New()
)

// loadEnvironmentSettings Loads settings specified as environment variables, or assigns default values
//
// Parameters:
//   - logger: Logger to be used
//
// Returns:
func loadEnvironmentSettings(logger log.Logger) {
	Environment = crosscutting.GetEnvironmentValue(logger, environmentEnv, "PROD")
	ApplicationMode = crosscutting.GetEnvironmentValue(logger, applicationModeEnv, "")
	if !(ApplicationMode == transmitterMode || ApplicationMode == receiverMode) {
		logger.Fatal(nil, "APPLICATION_MODE not found, please set Environment Variable: ["+applicationModeEnv+"], as ["+transmitterMode+"] or ["+receiverMode+"] ", "Configuration", "loadEnvironmentSettings")
	}

	if Environment != "PROD" {
		setupDevEnvironment(logger)
	} else {
		logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, loggingLevelEnv, "WARNING"))
		ClientID = crosscutting.GetEnvironmentValue(logger, serverOrgIDEnv, "")
		ServerURL = crosscutting.GetEnvironmentValue(logger, proxyURL, "http://localhost:8082")
	}

	if ClientID == "" {
		logger.Fatal(nil, "ClientID not found, please set Environment Variable: ["+serverOrgIDEnv+"]", "Configuration", "loadEnvironmentSettings")
	}
}

// setupDevEnvironment Sets up configuration values expected for development environment
//
// Parameters:
//   - logger: Logger to be used
//
// Returns:
func setupDevEnvironment(logger log.Logger) {
	logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, loggingLevelEnv, "DEBUG"))
	ClientID = crosscutting.GetEnvironmentValue(logger, serverOrgIDEnv, "09b20d09-bf30-4497-938e-b0ead8ce9629")
	ServerURL = crosscutting.GetEnvironmentValue(logger, proxyURL, "http://localhost:8082")
}

// Initialize Loads all settings requered for the application to run, such as endpoint settings and environment settings
//
// Parameters:
//
// Returns:
func Initialize() {
	logger := log.GetLogger()
	loadEnvironmentSettings(logger)
}
