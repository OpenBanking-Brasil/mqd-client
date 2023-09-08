package configuration

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
)

// Message contains the information of the payload to be validated
type EndPointSettings struct {
	Endpoint        string `json:"endpoint"`         // Name of the endpoint requested
	API             string `json:"api"`              // API name
	Group           string `json:"group"`            // API Group name
	ValidationRules string `json:"validation_rules"` // Path for the validation rules schema
	JSONSchema      string `json:"-"`
}

const ENDPOINT_SETTINGS_PATH = "ParameterData//endpoint_settings.json" // Constant to store path to the endpoint settings file.
const SERVER_ID_ENVIRONMENT = "SERVER_ORG_ID"                          //constant  to store name of the server id environment variable
const REPORT_EXECUTION_WINDOW = "REPORT_EXECUTION_WINDOW"              //constant  to store name of the report execution time environment variable

var (
	endPointSettings         []EndPointSettings // Buffered channel for message queue
	ServerId                 = ""               // Organisation id for server
	ReportExecutiontimeFrame = 0
)

func loadEndpointSettings() error {
	data, err := os.ReadFile(ENDPOINT_SETTINGS_PATH)
	if err != nil {
		log.Fatalf("error reading file: " + err.Error())
		return err
	}
	err = json.Unmarshal(data, &endPointSettings)
	if err != nil {
		log.Fatalf("error unmarshal file")
		return err
	}

	return nil
}

func loadEndpointValidationSchemas() error {
	for index := range endPointSettings {
		file, err := os.ReadFile(endPointSettings[index].ValidationRules)
		if err != nil {
			println(err.Error())
			return err
		}

		endPointSettings[index].JSONSchema = string(file)
	}

	return nil
}

func loadEnvironmentSettings() {
	ServerId = crosscutting.GetEnvironmentValue(SERVER_ID_ENVIRONMENT, "d2c118b2-1017-4857-a417-b0a346fdc5cc")
	crosscutting.GetEnvironmentValue(REPORT_EXECUTION_WINDOW, "1")
	intVar, err := strconv.Atoi(crosscutting.GetEnvironmentValue(REPORT_EXECUTION_WINDOW, "1"))
	if err != nil {
		intVar = 30
		println("REPORT_EXECUTION_WINDOW: Bad Format, Loading default: 30.")
	}

	ReportExecutiontimeFrame = intVar
}

func Initialize() error {
	loadEndpointSettings()
	loadEndpointValidationSchemas()
	loadEnvironmentSettings()
	return nil
}

func GetEndpointSettings() []EndPointSettings {
	return endPointSettings
}
