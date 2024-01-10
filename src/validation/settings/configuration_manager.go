package settings

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/server"
)

const api_configuration_name = "api_configuration.json"

var (
	singleton *ConfigurationManager // Singleton for configuration management
	mutex     = sync.Mutex{}        // Mutex for multi processing locks
)

type ConfigurationManager struct {
	pack             string            // Package name
	logger           log.Logger        // Logger to be used by the package
	apiGroupSettings *APIGroupSettings // Settings for each endpoint
	processRunning   bool              // Indicates that the process is running
}

func NewConfigurationManager(logger log.Logger) *ConfigurationManager {
	if singleton == nil {
		singleton = &ConfigurationManager{
			pack:   "settings.ConfigurationManager",
			logger: logger,
		}
	}

	return singleton
}

// getAPIConfigurationFile returns configuration settings for the specified parameters
func (cm *ConfigurationManager) getAPIConfigurationFile(basePath string, apiPath string, apiVersion string, server server.MQDServer) ([]APIEndpointSetting, error) {
	apiConfigurationpath := basePath + "//" + apiPath + "//" + apiVersion + "//response//"
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "ParameterData//", "")
	apiConfigurationpath = strings.ReplaceAll(apiConfigurationpath, "//", "/")
	fileName := apiConfigurationpath + "endpoints.json"
	cm.logger.Debug("loading File Name: "+fileName, cm.pack, "getAPIConfigurationFile")
	file, err := server.LoadAPIConfigurationFile(fileName)
	if err != nil {
		cm.logger.Error(err, "Error Reading Header schema file: "+fileName, cm.pack, "getAPIConfigurationFile")
		return nil, err
	}

	var result []APIEndpointSetting
	err = json.Unmarshal(file, &result)
	if err != nil {
		cm.logger.Error(err, "error unmarshal file", cm.pack, "getAPIConfigurationFile")
		return nil, err
	}

	return result, nil
}

// compareSchemaConfiguration compares the old configuration wiith a new one
// @author AB
// @params
// newSttings: new settings loaded from server
// server: Server to request the updates (if any)
// @return
// error: error if any
func (cm *ConfigurationManager) compareSchemaConfiguration(newSttings *APIGroupSettings, server server.MQDServer) (bool, error) {
	cm.logger.Info("Comparing Schemas", cm.pack, "compareSchemaConfiguration")
	schemaUpdated := false

	if cm.apiGroupSettings != nil {
		for i, newSet := range newSttings.Settings {
			cm.logger.Debug("Cehecking group: "+newSet.Group, cm.pack, "compareSchemaConfiguration")
			oldSet := cm.apiGroupSettings.GetGroupSetting(newSet.Group)
			if oldSet == nil {
				for j, newAPI := range newSet.ApiList {
					schemaUpdated = true
					epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version, server)
					if err != nil {
						cm.logger.Error(err, "error loading api configuration file", cm.pack, "compareSchemaConfiguration")
						return false, err
					}

					newSttings.Settings[i].ApiList[j].EndpointList = epList
				}
			} else {
				for j, newAPI := range newSet.ApiList {
					cm.logger.Debug("Cehecking API: "+newAPI.API, cm.pack, "compareSchemaConfiguration")
					oldAPI := oldSet.GetAPISetting(newAPI.API)
					if oldAPI == nil || oldAPI.Version != newAPI.Version {
						schemaUpdated = true
						cm.logger.Info("Updating API: "+newAPI.API, cm.pack, "compareSchemaConfiguration")
						epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version, server)
						if err != nil {
							cm.logger.Error(err, "error loading api configuration file", cm.pack, "compareSchemaConfiguration")
							return false, err
						}

						newSttings.Settings[i].ApiList[j].EndpointList = epList
					}
				}
			}
		}
	} else {
		cm.logger.Info("Executing first load", cm.pack, "compareSchemaConfiguration")
		for i, newSet := range newSttings.Settings {
			for j, newAPI := range newSet.ApiList {
				schemaUpdated = true
				cm.logger.Info("Loading API: "+newAPI.API, cm.pack, "compareSchemaConfiguration")
				epList, err := cm.getAPIConfigurationFile(newSet.BasePath, newAPI.BasePath, newAPI.Version, server)
				if err != nil {
					return false, err
				}

				newSttings.Settings[i].ApiList[j].EndpointList = epList
			}
		}
	}

	return schemaUpdated, nil
}

// updateValidationSchemas checks and updates the validation schemas for the endpoints
// @author AB
// @params
// @return
// error: error if any
func (cm *ConfigurationManager) updateValidationSchemas() error {
	cm.logger.Info("Updating Validation Schemas.", cm.pack, "updateValidationSchemas")
	cm.logger.Debug("Settings Path: "+api_configuration_name, "validation-settings", "LoadAPIConfigurationSettings")
	srv := server.NewMQDServer(cm.logger)

	// Load new schema map
	data, err := srv.LoadAPIConfigurationFile(api_configuration_name)
	if err != nil {
		cm.logger.Error(err, "Error loading API configuration", cm.pack, "updateValidationSchemas")
		return err
	}

	var tmpSettings APIGroupSettings
	err = json.Unmarshal(data, &tmpSettings.Settings)
	if err != nil {
		cm.logger.Error(err, "error unmarshal file", cm.pack, "updateValidationSchemas")
		cm.logger.Debug("Body: "+string(data), cm.pack, "updateValidationSchemas")
		return err
	}

	schemaUpdated, err := cm.compareSchemaConfiguration(&tmpSettings, *srv)
	if err != nil {
		cm.logger.Error(err, "Error Comparing API Schemas", cm.pack, "updateValidationSchemas")
		return err
	}

	mutex.Lock()
	if schemaUpdated {
		cm.apiGroupSettings = &tmpSettings
		configuration.LastUpdatedDate = time.Now()
	}

	mutex.Unlock()
	return nil
}

// updateConfiguration updates all configuration settings of the application
// @author AB
// @params
// @return
// error: error if any
func (cm *ConfigurationManager) updateConfiguration() error {
	cm.logger.Info("Executing configuration update", cm.pack, "updateConfiguration")
	return cm.updateValidationSchemas()
}

// getApiGroupSettings return the settings of API groups
// @author AB
// @params
// @return
// array of  APIGroupSetting found
func (cm *ConfigurationManager) getApiGroupSettings() []APIGroupSetting {
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()

	result := cm.apiGroupSettings.Settings
	return result
}

// StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
// @author AB
// @params
// @return
func (cm *ConfigurationManager) StartUpdateProcess() {
	if cm.processRunning {
		return
	}

	cm.processRunning = true
	cm.logger.Info("Starting configuration update Process", cm.pack, "StartUpdateProcess")
	timeWindow := time.Duration(6) * time.Hour

	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			cm.updateConfiguration()
		}
	}
}

// Initialize starts the initial settings
func (cm *ConfigurationManager) Initialize() error {
	return cm.updateConfiguration()
}

// getEndpointSettings loads a specific endpoint setting based on the endpoint name
// @author AB
// @params
// endpointName: Name of the endpoint to lookup for settings
// @return
// EndPointSettings: settings found, empty if no endpoint found
func (cm *ConfigurationManager) GetEndpointSettingFromAPI(endpointName string, logger log.Logger) *APIEndpointSetting {
	cm.logger.Info("loading Settings from API", cm.pack, "GetEndpointSettingFromAPI")
	settings := cm.getApiGroupSettings()

	for _, setting := range settings {
		for _, api := range setting.ApiList {
			if strings.Contains(strings.ToLower(endpointName), strings.ToLower(strings.TrimSpace(api.EndpointBase))) {
				for _, endpoint := range api.EndpointList {
					apiEndpointName := strings.ToLower(strings.TrimSpace(strings.TrimSpace(api.EndpointBase) + strings.TrimSpace(endpoint.Endpoint)))
					if apiEndpointName == strings.ToLower(strings.TrimSpace(endpointName)) {
						return &endpoint
					}
				}
			}
		}
	}

	logger.Debug("Endpoint Name not found.", "validation-settings", "GetEndpointSettingFromAPI")
	return nil
}
