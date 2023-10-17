package configuration

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// Message contains the information of the payload to be validated
type EndPointSettings struct {
	Endpoint              string `json:"endpoint"`                // Name of the endpoint requested
	API                   string `json:"api"`                     // API name
	Group                 string `json:"group"`                   // API Group name
	BasePath              string `json:"base_path"`               // Base path for the folder
	HeaderValidationRules string `json:"header_validation_rules"` // Header validation rules
	BodyValidationRules   string `json:"body_validation_rules"`   // Body validation rules
	JSONHeaderSchema      string `json:"-"`                       // Schema for the Header
	JSONBodySchema        string `json:"-"`                       // JSON schema for the Body
}

const ENDPOINT_SETTINGS_PATH = "ParameterData//endpoint_settings.json" // Constant to store path to the endpoint settings file.
const SERVER_ID_ENVIRONMENT = "SERVER_ORG_ID"                          //constant  to store name of the server id environment variable
const REPORT_EXECUTION_WINDOW = "REPORT_EXECUTION_WINDOW"              //constant  to store name of the report execution time environment variable
const LOGGING_LEVEL = "LOGGING_LEVEL"                                  //constant  to store name of the Logging level environment variable

var (
	endPointSettings         []EndPointSettings // Buffered channel for message queue
	ServerId                 = ""               // Organisation id for server
	ClientID                 = ""               // Organisation id for the client
	ReportExecutiontimeFrame = 0                // TimeWindow for report execution
)

// Func: loadEndpointSettings Loads the specific settings for each endpoint
// @author AB
// @params
// @return
// error: returns error in case of FILE_NOT_FOUND, or parsing error
func loadEndpointSettings() error {
	data, err := os.ReadFile(ENDPOINT_SETTINGS_PATH)
	if err != nil {
		log.Fatal(err, "error reading file", "Configuration", "loadEndpointSettings")
		return err
	}
	err = json.Unmarshal(data, &endPointSettings)
	if err != nil {
		log.Fatal(err, "error unmarshal file", "Configuration", "loadEndpointSettings")
		return err
	}

	return nil
}

// Func: loadEndpointValidationSchemas loads the Validation schemas from the file specified on the settings for each endpoint
// @author AB
// @params
// @return
// error: returns error in case of FILE_NOT_FOUND, or parsing error
func loadEndpointValidationSchemas() error {
	for index := range endPointSettings {
		file, err := os.ReadFile(endPointSettings[index].BasePath + endPointSettings[index].HeaderValidationRules)
		if err != nil {
			log.Error(err, "Error Reading Header schema file", "Configuration", "loadEndpointValidationSchemas")
			return err
		}

		endPointSettings[index].JSONHeaderSchema = string(file)

		file, err = os.ReadFile(endPointSettings[index].BasePath + endPointSettings[index].BodyValidationRules)
		if err != nil {
			log.Error(err, "Error Reading Body schema file", "Configuration", "loadEndpointValidationSchemas")
			return err
		}

		endPointSettings[index].JSONBodySchema = string(file)
	}

	return nil
}

// Func: loadEnvironmentSettings Loads settings specified as environment variables, or assigns default values
// @author AB
// @params
// @return
func loadEnvironmentSettings() {
	// ServerId = crosscutting.GetEnvironmentValue(SERVER_ID_ENVIRONMENT, "d2c118b2-1017-4857-a417-b0a346fdc5cc")
	ClientID = crosscutting.GetEnvironmentValue(SERVER_ID_ENVIRONMENT, "09b20d09-bf30-4497-938e-b0ead8ce9629")
	intVar, err := strconv.Atoi(crosscutting.GetEnvironmentValue(REPORT_EXECUTION_WINDOW, "10"))
	if err != nil {
		intVar = 30
		log.Log("REPORT_EXECUTION_WINDOW: Bad Format, Loading default: 30.", "Configuration", "loadEnvironmentSettings")
	}

	ReportExecutiontimeFrame = intVar

	log.SetLoggingGlobalLevelFromString(crosscutting.GetEnvironmentValue(LOGGING_LEVEL, "INFO"))
}

// Func: Initialize Loads all settings requered for the application to run, such as endpoint settings and environment settings
// @author AB
// @params
// @return
// error in case of load error.
func Initialize() error {
	loadEndpointSettings()
	loadEndpointValidationSchemas()
	loadEnvironmentSettings()
	return nil
}

// Func: GetEndpointSettings returns the lists of endpoint settings
// @author AB
// @params
// @return
// error in case of load error.
func GetEndpointSettings() []EndPointSettings {
	return endPointSettings
}

// Func: getEndpointSettings loads a specific endpoint setting based on the endpoint name
// @author AB
// @params
// endpointName: Name of the endpoint to lookup for settings
// @return
// EndPointSettings: settings found, empty if no endpoint found
func GetEndpointSetting(endpointName string) *EndPointSettings {
	for _, element := range GetEndpointSettings() {
		if element.Endpoint == endpointName {
			return &element
		}
	}

	return &EndPointSettings{}
}
