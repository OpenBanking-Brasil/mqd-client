package configuration

import (
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

const SERVER_ID_ENVIRONMENT = "SERVER_ORG_ID" // constant  to store name of the server id environment variable
const LOGGING_LEVEL = "LOGGING_LEVEL"         // constant  to store name of the Logging level environment variable
const ENVIRONMENT = "ENVIRONMENT"             // constant  to store name of the environment variable
const APPLICATION_MODE = "APPLICATION_MODE"   // constant  to store name of the application mode environment variable"
const TRANSMITTER_MODE = "TRANSMITTER"        // TRANSMITTER Application mode Constant
const RECEIVER_MODE = "RECEIVER"              // RECEIVER Application mode Constant

var (
	ServerId                = ""          // Organisation id for server
	ClientID                = ""          // Organisation id for the client
	Environment             = ""          // Indicates the actual Environment the app is running
	ApplicationMode         = ""          // Indicates the actual Application Mode the app is running in (TRANSMISSOR, RECEPTOR)
	LastReportExecutionDate = time.Time{} // Indicates the data of the last report execution
	ServerURL               = ""          // Server URL to send the reports to
	LastUpdatedDate         = time.Time{} // Indicates the data of the last report update
)

// loadEnvironmentSettings Loads settings specified as environment variables, or assigns default values
// @author AB
// @params
// @return
func loadEnvironmentSettings(logger log.Logger) {
	Environment = crosscutting.GetEnvironmentValue(logger, ENVIRONMENT, "PROD")
	ApplicationMode = crosscutting.GetEnvironmentValue(logger, APPLICATION_MODE, "")
	if !(ApplicationMode == TRANSMITTER_MODE || ApplicationMode == RECEIVER_MODE) {
		logger.Fatal(nil, "APPLICATION_MODE not found, please set Environment Variable: ["+APPLICATION_MODE+"], as ["+TRANSMITTER_MODE+"] or ["+RECEIVER_MODE+"] ", "Configuration", "loadEnvironmentSettings")
	}

	if Environment != "PROD" {
		setupDevEnvironment(logger)
	} else {
		logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, LOGGING_LEVEL, "WARNING"))
		ClientID = crosscutting.GetEnvironmentValue(logger, SERVER_ID_ENVIRONMENT, "")
		ServerURL = "https://mqd.openfinancebrasil.org.br"
	}

	if ClientID == "" {
		logger.Fatal(nil, "ClientID not found, please set Environment Variable: ["+SERVER_ID_ENVIRONMENT+"]", "Configuration", "loadEnvironmentSettings")
	}
}

// setupDevEnvironment Sets up configuration values expected for development environment
// @author AB
// @params
// @return
func setupDevEnvironment(logger log.Logger) {
	logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, LOGGING_LEVEL, "DEBUG"))
	ClientID = crosscutting.GetEnvironmentValue(logger, SERVER_ID_ENVIRONMENT, "09b20d09-bf30-4497-938e-b0ead8ce9629")
	// ServerURL = "https://auth-gateway-dev.openfinancebrasil.net.br"
	ServerURL = "http://localhost:8082"
}

// Initialize Loads all settings requered for the application to run, such as endpoint settings and environment settings
// @author AB
// @params
// @return
// error in case of load error.
func Initialize() {
	logger := log.GetLogger()
	loadEnvironmentSettings(logger)
}
