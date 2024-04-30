package crosscutting

import (
	"os"
	"strings"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// GetEnvironmentValue is obtaining a environment variable value
// @author AB
// @params
// key: Environment variable name
// defaultValue: Value to be used in case the variable is not asigned
// @return
func GetEnvironmentValue(logger log.Logger, key string, defaultValue string) string {
	result, found := os.LookupEnv(key)
	if !found {
		logger.Debug("Evironment Variable: ["+key+"], not found. using default value: ["+defaultValue+"]", "configuration", "GetEnvironmentValue")
		result = defaultValue
	}

	logger.Log("Evironment Variable: ["+key+"] = ["+result+"]", "configuration", "GetEnvironmentValue")
	return strings.TrimSpace(result)
}

// GetWorkingFolder returns the actual working folder of the application
// @author AB
// @params
// @return
// string Working folder
func GetWorkingFolder(logger log.Logger) string {
	path, err := os.Getwd()
	if err != nil {
		logger.Error(err, "Error getting working folder", "Crosscutting", "GetWorkingFolder")
	}

	return path
}
