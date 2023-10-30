package log

import "sync"

var (
	singleton Logger          // Logger variable to be used as a singleton
	lock      = &sync.Mutex{} // mutex for multithreading
)

// GetLogger returns the logger
// @author AB
// @param
// @return
// Logger created
func GetLogger() Logger {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()
		singleton = GetNewJSONLogger()
	}

	return singleton
}
