package settings

import (
	"encoding/json"

	"os"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// Message contains the information of the payload to be validated
type EndPointSetting struct {
	Endpoint              string `json:"endpoint"`                // Name of the endpoint requested
	API                   string `json:"api"`                     // API name
	Group                 string `json:"group"`                   // API Group name
	BasePath              string `json:"base_path"`               // Base path for the folder
	HeaderValidationRules string `json:"header_validation_rules"` // Header validation rules
	BodyValidationRules   string `json:"body_validation_rules"`   // Body validation rules
	JSONHeaderSchema      string `json:"-"`                       // Schema for the Header
	JSONBodySchema        string `json:"-"`                       // JSON schema for the Body
}

// Structure to store endpoint configuration settings
type EndPointSettings struct {
	Settings []EndPointSetting // Settings for each endpoint
}

var (
	endPointSettings EndPointSettings // Settings for each endpoint
)

// loadEndpointSettings Loads the specific settings for each endpoint
// @author AB
// @params
// @return
// error: returns error in case of FILE_NOT_FOUND, or parsing error
func LoadEndpointSettings(logger log.Logger, settingsPath string) error {
	logger.Info("Loading Endpoint settings.", "validation-settings", "LoadEndpointSettings")
	logger.Debug("Settings Path: "+settingsPath, "validation-settings", "LoadEndpointSettings")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		logger.Fatal(err, "error reading file", "validation-settings", "loadEndpointSettings")
		return err
	}
	err = json.Unmarshal(data, &endPointSettings.Settings)
	if err != nil {
		logger.Fatal(err, "error unmarshal file", "validation-settings", "loadEndpointSettings")
		return err
	}

	err = loadEndpointValidationSchemas(logger, endPointSettings)
	if err != nil {
		logger.Fatal(err, "Error loading validation schemas", "validation-settings", "loadEndpointSettings")
		return err
	}

	return nil
}

// GetEndPointSettings Returns the settings for each endpoint
// @author AB
// @params
// @return
// EndPointSettings: settings found, empty if no endpoint found
func GetEndPointSettings() *EndPointSettings {
	return &endPointSettings
}

// getEndpointSettings loads a specific endpoint setting based on the endpoint name
// @author AB
// @params
// endpointName: Name of the endpoint to lookup for settings
// @return
// EndPointSettings: settings found, empty if no endpoint found
func GetEndpointSetting(endpointName string) *EndPointSetting {
	settings := GetEndPointSettings().Settings
	for _, element := range settings {
		if element.Endpoint == endpointName {
			return &element
		}
	}

	return nil
}

// loadEndpointValidationSchemas loads the Validation schemas from the file specified on the settings for each endpoint
// @author AB
// @params
// @return
// error: returns error in case of FILE_NOT_FOUND, or parsing error
func loadEndpointValidationSchemas(logger log.Logger, endPointSettings EndPointSettings) error {
	logger.Info("Loading Endpoint Schemas.", "validation-settings", "loadEndpointValidationSchemas")
	for index := range endPointSettings.Settings {
		fileName := endPointSettings.Settings[index].BasePath + endPointSettings.Settings[index].HeaderValidationRules
		logger.Debug("Loading Header schema file: "+fileName, "validation-settings", "loadEndpointValidationSchemas")
		file, err := os.ReadFile(fileName)
		if err != nil {
			logger.Error(err, "Error Reading Header schema file: "+fileName, "validation-settings", "loadEndpointValidationSchemas")
			return err
		}

		endPointSettings.Settings[index].JSONHeaderSchema = string(file)

		fileName = endPointSettings.Settings[index].BasePath + endPointSettings.Settings[index].BodyValidationRules
		logger.Debug("Loading Body schema file: "+fileName, "validation-settings", "loadEndpointValidationSchemas")
		file, err = os.ReadFile(fileName)
		if err != nil {
			logger.Error(err, "Error Reading Body schema file", "Configuration", "loadEndpointValidationSchemas")
			return err
		}

		endPointSettings.Settings[index].JSONBodySchema = string(file)
	}

	return nil
}
