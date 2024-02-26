package application

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

const reportExecutionWindowEnv = "REPORT_EXECUTION_WINDOW" // constant  to store name of the report execution time environment variable
const reportExecutionNumberEnv = "REPORT_EXECUTION_NUMBER" // constant  to store name of the report execution total of reports environment variable

var (
	configurationManagerSingleton *ConfigurationManager // Singleton for configuration management
	configurationManagerMutex     = sync.Mutex{}        // Mutex for multi processing locks
)

// ConfigurationUpdateStatus stores the information of the configuration update process
type ConfigurationUpdateStatus struct {
	LastExecutionDate time.Time            // Indicates the data execution of the configuration update
	LastUpdatedDate   time.Time            // Indicates the data of the las succesful configuration update
	UpdateMessages    map[time.Time]string // List of error messages if any durin the update process
}

// ConfigurationManager is the manager in charge of handling configuration parameters of the application
type ConfigurationManager struct {
	crosscutting.OFBStruct
	ConfigurationSettings     *models.ConfigurationSettings // Configuration settings for the application
	processRunning            bool                          // Indicates that the process is running
	mqdServer                 services.ReportServer         // Report server for MQD
	configurationUpdateStatus ConfigurationUpdateStatus     // Last status of the configuration update
	environment               string
	reportExecutionWindow     int
	reportExecutionNumber     int
}

// NewConfigurationManager creates a new configuration manager for the application
//
// Parameters:
//   - logger: logger to be used
//   - mqdServer: MQD server to read the configuration files
//   - environment: Current configured environment of the application
//
// Returns:
//   - ConfigurationManager: new created configuration manager
func NewConfigurationManager(logger log.Logger, mqdServer services.ReportServer, environment string) *ConfigurationManager {
	if configurationManagerSingleton == nil {
		configurationManagerSingleton = &ConfigurationManager{
			OFBStruct: crosscutting.OFBStruct{
				Pack:   "application.ConfigurationManager",
				Logger: logger,
			},

			mqdServer:   mqdServer,
			environment: environment,
		}

		reportExecutionWindow := crosscutting.GetEnvironmentValue(logger, reportExecutionWindowEnv, "")
		if reportExecutionWindow != "" {
			value, err := strconv.Atoi(reportExecutionWindow)
			if err != nil {
				logger.Fatal(errors.New("report execution window environment variable unvalid character"), "Error loading REPORT_EXECUTION_WINDOW", configurationManagerSingleton.Pack, "NewConfigurationManager")
			}

			configurationManagerSingleton.reportExecutionWindow = value
		} else {
			configurationManagerSingleton.reportExecutionWindow = 0
		}

		reportExecutionNumber := crosscutting.GetEnvironmentValue(logger, reportExecutionNumberEnv, "")
		if reportExecutionNumber != "" {
			value, err := strconv.Atoi(reportExecutionWindow)
			if err != nil {
				logger.Fatal(errors.New("report execution Number environment variable unvalid character"), "Error loading REPORT_EXECUTION_NUMBER", configurationManagerSingleton.Pack, "NewConfigurationManager")
			}

			configurationManagerSingleton.reportExecutionNumber = value
		} else {
			configurationManagerSingleton.reportExecutionNumber = 0
		}

		configurationManagerSingleton.configurationUpdateStatus.UpdateMessages = make(map[time.Time]string)
	}

	return configurationManagerSingleton
}

// getAPIConfigurationFile returns configuration settings for the specified parameters
//
// Parameters:
//   - basePath: Base path of the api group
//   - apiPath: Path for the specific API
//   - apiVersion: api version of the endpoint
//
// Returns:
//   - []models.APIEndpointSetting: Array with endpoint settings for each of the endpoints in the api
//   - error: error if any
func (cm *ConfigurationManager) getAPIConfigurationFile(basePath string, apiPath string, apiVersion string) ([]models.APIEndpointSetting, error) {
	apiConfigurationpath := basePath + "//" + apiPath + "//" + apiVersion + "//response//"
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "ParameterData//", "")
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "//", "/")
	fileName := apiConfigurationpath + "endpoints.json"
	cm.Logger.Debug("loading File Name: "+fileName, cm.Pack, "getAPIConfigurationFile")
	file, err := cm.mqdServer.LoadAPIConfigurationFile(fileName)
	if err != nil {
		cm.Logger.Error(err, "Error Reading Header schema file: "+fileName, cm.Pack, "getAPIConfigurationFile")
		return nil, err
	}

	var result []models.APIEndpointSetting
	err = json.Unmarshal(file, &result)
	if err != nil {
		cm.Logger.Error(err, "error unmarshal file", cm.Pack, "getAPIConfigurationFile")
		return nil, err
	}

	return result, nil
}

// updateValidationSchemas checks and updates the validation schemas for the endpoints
//
// Parameters:
//   - newSettings: new configuration settings to update
//
// Returns:
//   - error: error if any
func (cm *ConfigurationManager) updateValidationSettings(newSettings *models.ConfigurationSettings) error {
	cm.Logger.Info("Updating Validation Schemas.", cm.Pack, "updateValidationSchemas")

	if cm.ConfigurationSettings == nil {
		cm.Logger.Info("Executing first load", cm.Pack, "updateValidationSettings")
		for i, newSet := range newSettings.ValidationSettings.APIGroupSettings {
			for j, newAPI := range newSet.APIList {
				cm.Logger.Info("Loading API: "+newAPI.API, cm.Pack, "updateValidationSettings")
				epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
				if err != nil {
					return err
				}

				newSettings.ValidationSettings.APIGroupSettings[i].APIList[j].EndpointList = epList
			}
		}

		return nil
	}

	for i, newSet := range newSettings.ValidationSettings.APIGroupSettings {
		oldSet := cm.ConfigurationSettings.ValidationSettings.GetGroupSetting(newSet.Group)
		if oldSet == nil {
			for j, newAPI := range newSet.APIList {
				epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
				if err != nil {
					cm.Logger.Error(err, "error loading api configuration file", cm.Pack, "updateValidationSettings")
					return err
				}

				newSettings.ValidationSettings.APIGroupSettings[i].APIList[j].EndpointList = epList
			}
		} else {
			for j, newAPI := range newSet.APIList {
				cm.Logger.Debug("Cehecking API: "+newAPI.API, cm.Pack, "updateValidationSettings")
				oldAPI := oldSet.GetAPISetting(newAPI.API)
				if oldAPI == nil || oldAPI.Version != newAPI.Version {
					cm.Logger.Info("Updating API: "+newAPI.API, cm.Pack, "updateValidationSettings")
					epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
					if err != nil {
						cm.Logger.Error(err, "error loading api configuration file", cm.Pack, "updateValidationSettings")
						return err
					}

					newSettings.ValidationSettings.APIGroupSettings[i].APIList[j].EndpointList = epList
				} else {
					newSettings.ValidationSettings.APIGroupSettings[i].APIList[j].EndpointList = oldAPI.EndpointList
				}
			}
		}
	}

	return nil
}

// updateConfiguration updates all configuration settings of the application
//
// Parameters:
//
// Returns:
//   - error: error if any
func (cm *ConfigurationManager) updateConfiguration() error {
	cm.Logger.Info("Executing configuration update", cm.Pack, "updateConfiguration")
	cm.configurationUpdateStatus.LastExecutionDate = time.Now()
	cs, err := cm.mqdServer.LoadConfigurationSettings()
	if err != nil {
		cm.configurationUpdateStatus.UpdateMessages[time.Now()] = err.Error()
		return err
	}

	if cm.ConfigurationSettings != nil && cs.Version == cm.ConfigurationSettings.Version {
		cm.Logger.Info("Same configuration version was found.", cm.Pack, "updateConfiguration")
		return nil
	}

	err = cm.updateValidationSettings(cs)
	if err != nil {
		cm.configurationUpdateStatus.UpdateMessages[cm.configurationUpdateStatus.LastExecutionDate] = err.Error()
		return err
	}

	configurationManagerMutex.Lock()
	cm.ConfigurationSettings = cs
	cm.configurationUpdateStatus.LastUpdatedDate = cm.configurationUpdateStatus.LastExecutionDate
	cm.configurationUpdateStatus.UpdateMessages = make(map[time.Time]string)
	cm.Logger.Info("Configuration was updated to the latest version: "+cm.ConfigurationSettings.Version, cm.Pack, "updateConfiguration")
	configurationManagerMutex.Unlock()

	return nil
}

// getAPIGroupSettings return the settings of API groups
//
// Parameters:
//
// Returns:
//   - []models.APIGroupSetting: Array of APIGroupSetting found
func (cm *ConfigurationManager) getAPIGroupSettings() []models.APIGroupSetting {
	configurationManagerMutex.Lock()
	defer func() {
		configurationManagerMutex.Unlock()
	}()

	result := cm.ConfigurationSettings.ValidationSettings.APIGroupSettings
	return result
}

// StartUpdateProcess starts the periodic process that prints total results and clears them every 2 minutes
//
// Parameters:
//
// Returns:
func (cm *ConfigurationManager) StartUpdateProcess() {
	if cm.processRunning {
		return
	}

	cm.processRunning = true
	cm.Logger.Info("Starting configuration update Process", cm.Pack, "StartUpdateProcess")
	timeWindow := time.Duration(2) * time.Minute
	if cm.environment != "DEBUG" {
		timeWindow = time.Duration(4) * time.Hour
	}

	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			cm.updateConfiguration()
		}
	}
}

// Initialize executes initial settings configuration
//
// Parameters:
//
// Returns:
//   - error: error if any
func (cm *ConfigurationManager) Initialize() error {
	return cm.updateConfiguration()
}

// GetEndpointSettingFromAPI loads a specific endpoint setting based on the endpoint name
//
// Parameters:
//   - endpointName: Name of the endpoint to lookup for settings
//   - logger: logger object to be used
//
// Returns:
//   - *models.APIEndpointSetting: error if any
//   - string: version of the api
func (cm *ConfigurationManager) GetEndpointSettingFromAPI(endpointName string, logger log.Logger) (*models.APIEndpointSetting, string) {
	cm.Logger.Info("loading Settings from API", cm.Pack, "GetEndpointSettingFromAPI")
	settings := cm.getAPIGroupSettings()

	for _, setting := range settings {
		for _, api := range setting.APIList {
			if strings.Contains(strings.ToLower(endpointName), strings.ToLower(strings.TrimSpace(api.EndpointBase))) {
				for _, endpoint := range api.EndpointList {
					apiEndpointName := strings.ToLower(strings.TrimSpace(strings.TrimSpace(api.EndpointBase) + strings.TrimSpace(endpoint.Endpoint)))
					if apiEndpointName == strings.ToLower(strings.TrimSpace(endpointName)) {
						return &endpoint, api.Version
					}
				}
			}
		}
	}

	logger.Debug("Endpoint Name not found.", "validation-settings", "GetEndpointSettingFromAPI")
	return nil, ""
}

// GetLastExecutionDate returns the las execution date
//
// Parameters:
//
// Returns:
//   - time.Time: Last execution time
func (cm *ConfigurationManager) GetLastExecutionDate() time.Time {
	return cm.configurationUpdateStatus.LastExecutionDate
}

// GetLastUpdatedDate returns the las updated date
//
// Parameters:
//
// Returns:
//   - time.Time: Last updated time
func (cm *ConfigurationManager) GetLastUpdatedDate() time.Time {
	return cm.configurationUpdateStatus.LastUpdatedDate
}

// GetUpdateMessages returns the list of update messages
//
// Parameters:
//
// Returns:
//   - map: map[time.Time]string with the list of messages by date
func (cm *ConfigurationManager) GetUpdateMessages() map[time.Time]string {
	return cm.configurationUpdateStatus.UpdateMessages
}

// GetReportExecutionWindow returns the report execution window configured
//
// Parameters:
//
// Returns:
//   - int: report execution window in minutes
func (cm *ConfigurationManager) GetReportExecutionWindow() int {
	if cm.reportExecutionWindow > 0 {
		return cm.reportExecutionWindow
	}

	return cm.ConfigurationSettings.ReportSettings.ReportExecutionWindow
}

// GetSendOnReportNumber returns the number of reports that should be sent
//
// Parameters:
//
// Returns:
//   - int: number of reports to check
func (cm *ConfigurationManager) GetSendOnReportNumber() int {
	if cm.reportExecutionNumber > 0 {
		return cm.reportExecutionNumber
	}

	return cm.ConfigurationSettings.ReportSettings.SendOnReportNumber
}
