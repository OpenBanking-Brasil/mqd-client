package configuration

import (
	"strconv"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/validation/settings"
)

const ENDPOINT_SETTINGS_PATH = "ParameterData//endpoint_settings.json" // Constant to store path to the endpoint settings file.
const SERVER_ID_ENVIRONMENT = "SERVER_ORG_ID"                          // constant  to store name of the server id environment variable
const REPORT_EXECUTION_WINDOW = "REPORT_EXECUTION_WINDOW"              // constant  to store name of the report execution time environment variable
const LOGGING_LEVEL = "LOGGING_LEVEL"                                  // constant  to store name of the Logging level environment variable
const ENVIRONMENT = "ENVIRONMENT"                                      // constant  to store name of the environment variable

var (
	ServerId                 = "" // Organisation id for server
	ClientID                 = "" // Organisation id for the client
	ReportExecutiontimeFrame = 0  // TimeWindow for report execution
	Environment              = "" // Indicates the actual Environment the app is running
	ServerURL                = "" // Server URL to send the reports to
)

// loadEnvironmentSettings Loads settings specified as environment variables, or assigns default values
// @author AB
// @params
// @return
func loadEnvironmentSettings(logger log.Logger) {
	Environment = crosscutting.GetEnvironmentValue(logger, ENVIRONMENT, "PROD")

	if Environment != "PROD" {
		setupDevEnvironment(logger)
	} else {
		logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, LOGGING_LEVEL, "WARNING"))
		intVar, err := strconv.Atoi(crosscutting.GetEnvironmentValue(logger, REPORT_EXECUTION_WINDOW, "10"))
		if err != nil {
			intVar = 30
			logger.Log("REPORT_EXECUTION_WINDOW: Bad Format, Loading default: 30.", "Configuration", "loadEnvironmentSettings")
		}

		ReportExecutiontimeFrame = intVar
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
	ReportExecutiontimeFrame = 1
	logger.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(logger, LOGGING_LEVEL, "DEBUG"))
	ClientID = crosscutting.GetEnvironmentValue(logger, SERVER_ID_ENVIRONMENT, "09b20d09-bf30-4497-938e-b0ead8ce9629")
	ServerURL = "https://auth-gateway-dev.openfinancebrasil.net.br"
}

// Initialize Loads all settings requered for the application to run, such as endpoint settings and environment settings
// @author AB
// @params
// @return
// error in case of load error.
func Initialize() {
	logger := log.GetLogger()
	loadEnvironmentSettings(logger)
	settings.LoadEndpointSettings(logger, ENDPOINT_SETTINGS_PATH)
}
