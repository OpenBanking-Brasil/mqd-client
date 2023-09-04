package crosscutting

import (
	"os"
)

/**
 * Func: GetEnvironmentValue is obtaining a environment variable value
 *
 * @author AB
 *
 * @params
 * key: Environment variable name
 * defaultValue: Value to be used in case the variable is not asigned
 * @return
 */
func GetEnvironmentValue(key string, defaultValue string) string {
	result := os.Getenv(key)
	if result == "" {
		result = defaultValue
	}

	return result
}

func GetWorkingFolder() string {
	path, err := os.Getwd()
	if err != nil {
		println(err)
	}

	return path
}
