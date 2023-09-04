package configuration

import (
	"encoding/json"
	"os"
)

// Message contains the information of the payload to be validated
type EndPointSettings struct {
	Endpoint        string `json:"endpoint"`         // Name of the endpoint requested
	API             string `json:"api"`              // API name
	Group           string `json:"group"`            // API Group name
	ValidationRules string `json:"validation_rules"` // Path for the validation rules schema
	JSONSchema      string `json:"-"`
}

// Buffered channel for message queue
var endPointSettings []EndPointSettings

func loadEndpointSettings() error {
	println("Loading settings")
	data, err := os.ReadFile("ParameterData\\endpoint_settings.json")
	if err != nil {
		println("error reading file: " + err.Error())
		return err
	}
	err = json.Unmarshal(data, &endPointSettings)
	if err != nil {
		println("error unmarshal file")
		return err
	}

	return nil
}

func loadEndpointValidationSchemas() error {
	println("Loading schemas")
	for index, _ := range endPointSettings {
		file, err := os.ReadFile(endPointSettings[index].ValidationRules)
		if err != nil {
			println(err.Error())
			return err
		}

		endPointSettings[index].JSONSchema = string(file)
	}

	return nil
}

func Initialize() error {
	loadEndpointSettings()
	loadEndpointValidationSchemas()
	return nil
}

func GetEndpointSettings() []EndPointSettings {
	return endPointSettings
}
