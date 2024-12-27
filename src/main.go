package main

import (
	"github.com/OpenBanking-Brasil/MQD_Client/application"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

var (
	logger   log.Logger
	settings configuration.Settings
)

func init() {
	monitoring.StartOpenTelemetry()
	cnf := configuration.Configuration{}
	settings = cnf.GetApplicationSettings()
	logger = log.GetLogger()
}

// Main is the main function of the api, that is executed on "run"
// @author AB
// @params
// @return
func main() {
	reportServer := services.GetReportServer(logger, settings.SecuritySettings.ProxyURL, settings)
	cm := application.NewConfigurationManager(logger, *reportServer, settings)
	err := cm.Initialize()
	if err != nil {
		logger.Fatal(err, "There was a fatal error loading initial settings.", "Main", "Main")
	}

	qm := application.GetQueueManager()
	rp := application.GetResultProcessor(logger, *reportServer, cm)
	lrm := application.NewLocalResultManager(logger, cm)
	mp := application.GetMessageProcessorWorker(logger, rp, qm, cm, lrm)

	// Start workers
	go cm.StartUpdateProcess()
	go mp.StartWorker()
	go rp.StartResultsProcessor()
	go lrm.StartResultProcess()

	application.GetAPIServer(logger, monitoring.GetOpentelemetryHandler(), qm, cm).StartServing()
}
