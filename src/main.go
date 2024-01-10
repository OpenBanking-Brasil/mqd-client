package main

import (
	"github.com/OpenBanking-Brasil/MQD_Client/apiserver"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/validation/settings"
	"github.com/OpenBanking-Brasil/MQD_Client/worker"
)

// Main is the main function of the api, that is executed on "run"
// @author AB
// @params
// @return
func main() {
	monitoring.StartOpenTelemetry()
	configuration.Initialize()
	logger := log.GetLogger()

	cm := settings.NewConfigurationManager(logger)
	err := cm.Initialize()
	if err != nil {
		logger.Fatal(err, "There was a fatal error loading initial settings.", "Main", "Main")
	}

	go cm.StartUpdateProcess()

	rp := result.GetResultProcessor(logger)
	mp := worker.GetMessageProcessorWorker(logger, rp, cm)

	// Start the worker Goroutine to process messages
	go mp.StartWorker()
	// go rp.StartResultsProcessor()

	apiserver.GetAPIServer(logger, monitoring.GetOpentelemetryHandler(), cm).StartServing()
}
