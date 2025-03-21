package log

import "sync"

var (
	singleton Logger          // Logger variable to be used as a singleton
	lock      = &sync.Mutex{} // mutex for multithreading
)

// GetLogger returns the logger
//
// Parameters:
//
// Returns:
//   - logger: Logger created
func GetLogger(loggingLevel string) Logger {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()
		singleton = GetNewJSONLogger()
		singleton.SetLoggingGlobalLevelFromString(loggingLevel)
	}

	return singleton
}
