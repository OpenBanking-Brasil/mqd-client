package main

import (
	"encoding/json"
	"io/fs"
	"os"

	"github.com/OpenBanking-Brasil/MQD_Client/application"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

// Main is the main function of the api, that is executed on "run"
// @author AB
// @params
// @return
func main() {
	monitoring.StartOpenTelemetry()
	configuration.Initialize()
	// monitoring.IncreaseBadEndpointsReceived("Test endpoint 1", "N.A.", "Endpoint not supported")
	// monitoring.IncreaseBadEndpointsReceived("Another Endpoint 1", "1.0.0", "Version not supported")

	// monitoring.IncreaseBadEndpointsReceived("Test endpoint 1", "N.A.", "Endpoint not supported")
	// monitoring.IncreaseBadEndpointsReceived("Another Endpoint", "1.0.0", "Version not supported")

	// monitoring.IncreaseBadEndpointsReceived("Test endpoint 2", "N.A.", "Endpoint not supported")

	logger := log.GetLogger()
	reportServer := services.GetReportServer(logger)

	cm := application.NewConfigurationManager(logger, *reportServer, configuration.Environment)
	err := cm.Initialize()
	if err != nil {
		logger.Fatal(err, "There was a fatal error loading initial settings.", "Main", "Main")
	}

	qm := application.GetQueueManager()
	rp := application.GetResultProcessor(logger, *reportServer, cm)
	mp := application.GetMessageProcessorWorker(logger, rp, qm, cm)

	// Start workers
	go cm.StartUpdateProcess()
	go mp.StartWorker()
	go rp.StartResultsProcessor()

	application.GetAPIServer(logger, monitoring.GetOpentelemetryHandler(), qm, cm).StartServing()
}

func createConfigurationSettingsFile() {
	cs := &models.ConfigurationSettings{
		Version: "1.0.0",
		ValidationSettings: models.ValidationSettings{
			TransmitterValidationRate: 100,
			ReceiverValidationRate:    20,
			APIGroupSettings: []models.APIGroupSetting{{
				Group:    "Dados Cadastrais e Transacionais",
				BasePath: "ParameterData//transactions",
				ApiList: []models.APISetting{{
					API:          "contas",
					BasePath:     "accounts",
					Version:      "2.2.0",
					EndpointBase: "/accounts/v2",
				}, {
					API:          "contas",
					BasePath:     "accounts",
					Version:      "2.2.0",
					EndpointBase: "/accounts/v2",
				},
				},
			}},
		},
		ReportSettings: models.ReportSettings{
			ReportExecutionWindow: 10,
			SendOnReportNumber:    1000000,
		},
	}

	WriteFile(cs, "configurationSettings.json")
}

func WriteFile(data interface{}, fileName string) {
	// serialize object to json
	jsonBytes, erro := json.MarshalIndent(data, "", "  ")
	if erro != nil {

	}

	erro = os.WriteFile(fileName, jsonBytes, fs.ModeAppend)
	if erro != nil {

	}
}
