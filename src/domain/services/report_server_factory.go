package services

import (
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

var (
	lock      = &sync.Mutex{} // mutex for multithreading
	singleton ReportServer    // Singleton for the Report Server
)

// Returns the report server to be used
// @author AB
// @params
// logger: the logger to be used
// @return
// ReportServer instance.
func GetReportServer(logger log.Logger) *ReportServer {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()
		singleton = NewReportServerMQD(logger)
	}

	return &singleton
}
