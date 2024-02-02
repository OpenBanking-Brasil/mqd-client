package application

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

// const api_configuration_name = "api_configuration.json"

var (
	configurationManagerSingleton *ConfigurationManager // Singleton for configuration management
	configurationManagerMutex     = sync.Mutex{}        // Mutex for multi processing locks
)

type ConfigurationUpdateStatus struct {
	LastExecutionDate time.Time            // Indicates the data execution of the configuration update
	LastUpdatedDate   time.Time            // Indicates the data of the las succesful configuration update
	UpdateMessages    map[time.Time]string // List of error messages if any durin the update process
}

type ConfigurationManager struct {
	crosscutting.OFBStruct
	ConfigurationSettings     *models.ConfigurationSettings // Configuration settings for the application
	processRunning            bool                          // Indicates that the process is running
	mqdServer                 services.ReportServer         // Report server for MQD
	configurationUpdateStatus ConfigurationUpdateStatus     // Last status of the configuration update
	environment               string
}

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

		configurationManagerSingleton.configurationUpdateStatus.UpdateMessages = make(map[time.Time]string)
	}

	return configurationManagerSingleton
}

// getAPIConfigurationFile returns configuration settings for the specified parameters
func (this *ConfigurationManager) getAPIConfigurationFile(basePath string, apiPath string, apiVersion string) ([]models.APIEndpointSetting, error) {
	apiConfigurationpath := basePath + "//" + apiPath + "//" + apiVersion + "//response//"
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "ParameterData//", "")
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "//", "/")
	fileName := apiConfigurationpath + "endpoints.json"
	this.Logger.Debug("loading File Name: "+fileName, this.Pack, "getAPIConfigurationFile")
	file, err := this.mqdServer.LoadAPIConfigurationFile(fileName)
	if err != nil {
		this.Logger.Error(err, "Error Reading Header schema file: "+fileName, this.Pack, "getAPIConfigurationFile")
		return nil, err
	}

	var result []models.APIEndpointSetting
	err = json.Unmarshal(file, &result)
	if err != nil {
		this.Logger.Error(err, "error unmarshal file", this.Pack, "getAPIConfigurationFile")
		return nil, err
	}

	return result, nil
}

// updateValidationSchemas checks and updates the validation schemas for the endpoints
// @author AB
// @params
// @return
// error: error if any
func (this *ConfigurationManager) updateValidationSettings(newSettings *models.ConfigurationSettings) error {
	this.Logger.Info("Updating Validation Schemas.", this.Pack, "updateValidationSchemas")

	if this.ConfigurationSettings == nil {
		this.Logger.Info("Executing first load", this.Pack, "updateValidationSettings")
		for i, newSet := range newSettings.ValidationSettings.APIGroupSettings {
			for j, newAPI := range newSet.ApiList {
				this.Logger.Info("Loading API: "+newAPI.API, this.Pack, "updateValidationSettings")
				epList, err := this.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
				if err != nil {
					return err
				}

				newSettings.ValidationSettings.APIGroupSettings[i].ApiList[j].EndpointList = epList
			}
		}

		return nil
	}

	for i, newSet := range newSettings.ValidationSettings.APIGroupSettings {
		oldSet := this.ConfigurationSettings.ValidationSettings.GetGroupSetting(newSet.Group)
		if oldSet == nil {
			for j, newAPI := range newSet.ApiList {
				epList, err := this.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
				if err != nil {
					this.Logger.Error(err, "error loading api configuration file", this.Pack, "updateValidationSettings")
					return err
				}

				newSettings.ValidationSettings.APIGroupSettings[i].ApiList[j].EndpointList = epList
			}
		} else {
			for j, newAPI := range newSet.ApiList {
				this.Logger.Debug("Cehecking API: "+newAPI.API, this.Pack, "updateValidationSettings")
				oldAPI := oldSet.GetAPISetting(newAPI.API)
				if oldAPI == nil || oldAPI.Version != newAPI.Version {
					this.Logger.Info("Updating API: "+newAPI.API, this.Pack, "updateValidationSettings")
					epList, err := this.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version)
					if err != nil {
						this.Logger.Error(err, "error loading api configuration file", this.Pack, "updateValidationSettings")
						return err
					}

					newSettings.ValidationSettings.APIGroupSettings[i].ApiList[j].EndpointList = epList
				} else {
					newSettings.ValidationSettings.APIGroupSettings[i].ApiList[j].EndpointList = oldAPI.EndpointList
				}
			}
		}
	}

	return nil
}

// updateConfiguration updates all configuration settings of the application
// @author AB
// @params
// @return
// error: error if any
func (this *ConfigurationManager) updateConfiguration() error {
	this.Logger.Info("Executing configuration update", this.Pack, "updateConfiguration")
	this.configurationUpdateStatus.LastExecutionDate = time.Now()
	cs, err := this.mqdServer.LoadConfigurationSettings()
	if err != nil {
		this.configurationUpdateStatus.UpdateMessages[time.Now()] = err.Error()
		return err
	}

	if this.ConfigurationSettings != nil && cs.Version == this.ConfigurationSettings.Version {
		this.Logger.Info("Same configuration version was found.", this.Pack, "updateConfiguration")
		return nil
	}

	err = this.updateValidationSettings(cs)
	if err != nil {
		this.configurationUpdateStatus.UpdateMessages[this.configurationUpdateStatus.LastExecutionDate] = err.Error()
		return err
	}

	configurationManagerMutex.Lock()
	this.ConfigurationSettings = cs
	this.configurationUpdateStatus.LastUpdatedDate = this.configurationUpdateStatus.LastExecutionDate
	this.configurationUpdateStatus.UpdateMessages = make(map[time.Time]string)
	this.Logger.Info("Configuration was updated to the latest version: "+this.ConfigurationSettings.Version, this.Pack, "updateConfiguration")
	configurationManagerMutex.Unlock()

	this.configurationUpdateStatus.UpdateMessages[time.Now()] = "Error loading file."
	this.configurationUpdateStatus.UpdateMessages[time.Now().Add(time.Duration(1)*time.Minute)] = "File Not found."

	return nil
}

// getApiGroupSettings return the settings of API groups
// @author AB
// @params
// @return
// array of  APIGroupSetting found
func (this *ConfigurationManager) getApiGroupSettings() []models.APIGroupSetting {
	configurationManagerMutex.Lock()
	defer func() {
		configurationManagerMutex.Unlock()
	}()

	result := this.ConfigurationSettings.ValidationSettings.APIGroupSettings
	return result
}

// StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
// @author AB
// @params
// @return
func (this *ConfigurationManager) StartUpdateProcess() {
	if this.processRunning {
		return
	}

	this.processRunning = true
	this.Logger.Info("Starting configuration update Process", this.Pack, "StartUpdateProcess")
	timeWindow := time.Duration(2) * time.Minute
	if this.environment != "DEBUG" {
		timeWindow = time.Duration(2) * time.Hour
	}

	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			this.updateConfiguration()
		}
	}
}

// Initialize starts the initial settings
func (this *ConfigurationManager) Initialize() error {
	return this.updateConfiguration()
}

// getEndpointSettings loads a specific endpoint setting based on the endpoint name
// @author AB
// @params
// endpointName: Name of the endpoint to lookup for settings
// @return
// EndPointSettings: settings found, empty if no endpoint found
func (this *ConfigurationManager) GetEndpointSettingFromAPI(endpointName string, logger log.Logger) (*models.APIEndpointSetting, string) {
	this.Logger.Info("loading Settings from API", this.Pack, "GetEndpointSettingFromAPI")
	settings := this.getApiGroupSettings()

	for _, setting := range settings {
		for _, api := range setting.ApiList {
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

func (this *ConfigurationManager) GetLastExecutionDate() time.Time {
	return this.configurationUpdateStatus.LastExecutionDate
}

func (this *ConfigurationManager) GetLastUpdatedDate() time.Time {
	return this.configurationUpdateStatus.LastUpdatedDate
}

func (this *ConfigurationManager) GetUpdateMessages() map[time.Time]string {
	return this.configurationUpdateStatus.UpdateMessages
}

func (this *ConfigurationManager) GetReportExecutionWindow() int {
	if this.environment == "DEBUG" {
		return 3
	}

	return this.ConfigurationSettings.ReportSettings.ReportExecutionWindow
}

func (this *ConfigurationManager) GetSendOnReportNumber() int {
	if this.environment == "DEBUG" {
		return 30
	}

	return this.ConfigurationSettings.ReportSettings.SendOnReportNumber
}
