package services

import (
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

var (
	lock      = &sync.Mutex{} // mutex for multithreading
	singleton ReportServer    // Singleton for the Report Server
)

// GetReportServer Returns the report server to be used
//
// Parameters:
//   - logger: Logger to be used
//
// Returns:
//   - ReportServer: ReportServer instance
func GetReportServer(logger log.Logger) *ReportServer {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()
		singleton = NewReportServerMQD(logger)
	}

	return &singleton
}
