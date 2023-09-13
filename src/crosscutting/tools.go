package crosscutting

import (
	"os"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// Func: GetEnvironmentValue is obtaining a environment variable value
// @author AB
// @params
// key: Environment variable name
// defaultValue: Value to be used in case the variable is not asigned
// @return
func GetEnvironmentValue(key string, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		result = defaultValue
	}

	return result
}

// Func: GetWorkingFolder returns the actual working folder of the application
// @author AB
// @params
// @return
// string Working folder
func GetWorkingFolder() string {
	path, err := os.Getwd()
	if err != nil {
		log.Error(err, "Error getting working folder", "Crosscutting", "GetWorkingFolder")
	}

	return path
}
