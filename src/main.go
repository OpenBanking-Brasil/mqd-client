package main

import (
	"github.com/OpenBanking-Brasil/MQD_Client/application"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

// Main is the main function of the api, that is executed on "run"
// @author AB
// @params
// @return
func main() {
	monitoring.StartOpenTelemetry()
	configuration.Initialize()
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
