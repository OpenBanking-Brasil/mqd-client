package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/validation"
)

const (
	resultTimeFormat = "2006-01-02"
)

var (
	localResultMutex = sync.Mutex{} // Mutex for thread-safe access to messageResults
	mu               = sync.Mutex{} // Mutex for thread-safe access to messageResults
	basePath         = "./data_logs"
)

type payloadDetail struct {
	XFapiInteractionID string
	ConsentID          string
	Payload            validation.DynamicStruct
	Errors             map[string][]string
}

type localEndpointSummary struct {
	EndpointName       string
	Requests           int
	RequestsWithErrors int
	PayloadDetails     []payloadDetail
}

// LocalResultManager is the manager in charge of handling local results
type LocalResultManager struct {
	crosscutting.OFBStruct
	cm             *ConfigurationManager // Manager for application settings
	result         map[string]localEndpointSummary
	recordedErrors map[string]int
	lstCleanupDate string
}

// NewLocalResultManager creates a new Local result manager
//
// Parameters:
//   - logger: logger to be used
//   - cm: Configuration manager to be used
//
// Returns:
//   - ConfigurationManager: new created Local result manager
func NewLocalResultManager(logger log.Logger, cm *ConfigurationManager) *LocalResultManager {
	return &LocalResultManager{
		OFBStruct: crosscutting.OFBStruct{
			Pack:   "application.LocalResultManager",
			Logger: logger,
		},
		cm:             cm,
		result:         make(map[string]localEndpointSummary),
		recordedErrors: make(map[string]int),
	}
}

// StartResultProcess Start the process of storage and cleanup of files
//
// Parameters:
//
// Returns:
func (mng *LocalResultManager) StartResultProcess() {
	go mng.startStoreProcess()
	go mng.startCleanupProcess()
}

// AppendResult Includes a result to be stored into the results file
//
// Parameters:
//   - message: Message to be stored
//   - result: Validation result of the message
//   - settings: API validation settings configured
//
// Returns:
func (mng *LocalResultManager) AppendResult(message Message, result MessageResult, settings APIValidationSettings) {
	if !mng.cm.settings.ResultSettings.Enabled {
		return
	}

	localResultMutex.Lock()

	key := fmt.Sprintf("%s-%s-%s", settings.APIGroup, strings.ReplaceAll(settings.BasePath, "-", ""), settings.EndpointSettings.Endpoint)
	if _, ok := mng.result[key]; !ok {
		mng.result[key] = localEndpointSummary{
			EndpointName:       settings.EndpointSettings.Endpoint,
			Requests:           0,
			RequestsWithErrors: 0,
			PayloadDetails:     make([]payloadDetail, 0),
		}
	}

	summary := mng.result[key]
	summary.Requests++

	if !result.Result {
		summary.RequestsWithErrors++
		needToSaveSample := false
		for field, errorField := range result.Errors {
			for _, validError := range errorField {
				errorKey := fmt.Sprintf("%s-%s-%s-%s-%s", settings.APIGroup, strings.ReplaceAll(settings.BasePath, "-", ""), settings.EndpointSettings.Endpoint, field, validError)
				if mng.recordedErrors[errorKey] >= mng.cm.settings.ResultSettings.SamplesPerError {
					continue
				} else {
					mng.recordedErrors[errorKey]++
					needToSaveSample = true
				}
			}
		}

		if needToSaveSample {
			payload, err := message.GetMappedObject()
			if err != nil {
				mng.Logger.Error(err, "there was an error while loading the message object", mng.Pack, "AppendResult")
			}

			payload = mng.findAndScrambleAttribute(payload)
			newDetail := payloadDetail{
				Payload:            payload,
				ConsentID:          message.ConsentID,
				XFapiInteractionID: message.XFapiInteractionID,
				Errors:             result.Errors,
			}
			summary.PayloadDetails = append(summary.PayloadDetails, newDetail)
		}
	}

	mng.result[key] = summary
	localResultMutex.Unlock()
}

func (mng *LocalResultManager) startStoreProcess() {
	if !mng.cm.settings.ResultSettings.Enabled {
		return
	}

	executionWindow := 24 / mng.cm.settings.ResultSettings.FilesPerDay
	if executionWindow == 0 {
		mng.Logger.Panic("FilesPerDay value is higher than expected, max value : 24, min value: 1.", mng.Pack, "StartStoreProcess")
	}

	timeWindow := time.Duration(executionWindow) * time.Hour
	ticker := time.NewTicker(timeWindow)

	for range ticker.C {
		mng.storeFiles()
	}
}

func (mng *LocalResultManager) storeFiles() {
	mng.Logger.Info("Executing  store log files.", mng.Pack, "startStoreProcess")
	if len(mng.result) <= 0 {
		return
	}

	localResultMutex.Lock()

	reports := mng.result
	mng.result = make(map[string]localEndpointSummary)
	mng.recordedErrors = make(map[string]int)
	localResultMutex.Unlock()
	filesToSave := make(map[string][]localEndpointSummary)
	for key, value := range reports {
		keyValues := strings.Split(key, "-")
		api := keyValues[1]

		filesToSave[api] = append(filesToSave[api], value)
	}

	for key, file := range filesToSave {
		err := mng.saveFile(basePath, mng.cm.settings.ConfigurationSettings.ApplicationID.String(), key, file)
		if err != nil {
			mng.Logger.Error(err, "there was an error saving data file", mng.Pack, "storeFiles")
		}
	}
}

func (mng *LocalResultManager) startCleanupProcess() {
	for {
		mu.Lock()
		today := time.Now().Format(resultTimeFormat)
		if mng.lstCleanupDate != today {
			// Update last run date and execute the task
			mng.lstCleanupDate = today
			mu.Unlock()
			mng.cleanupFiles()
		} else {
			mu.Unlock()
		}
		// Check once per hour to minimize CPU usage
		time.Sleep(1 * time.Hour)
	}
}

func (mng *LocalResultManager) cleanupFiles() {
	// Calculate the cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -mng.cm.settings.ResultSettings.DaysToStore)

	// Walk through the directory
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			mng.Logger.Error(err, "Error reading folder", mng.Pack, "cleanupFiles")
			return err
		}

		// Skip files and focus on directories
		if !info.IsDir() || path == basePath {
			return nil
		}

		// Parse the folder name as a date
		folderName := info.Name()
		folderDate, err := time.Parse(resultTimeFormat, folderName)
		if err != nil {
			// not a date folder, must skip
			return nil
		}

		// Check if the folder is older than the cutoff date
		if folderDate.Before(cutoffDate) {
			mng.Logger.Error(err, "Removing folder: "+path, mng.Pack, "cleanupFiles")
			return os.RemoveAll(path) // Remove the folder and its contents
		}

		return nil
	})

	if err != nil {
		mng.Logger.Error(err, "There was an error while reading folders", mng.Pack, "cleanupFiles")
	}
}

func (mng *LocalResultManager) saveFile(basePath string, appID string, familyType string, data []localEndpointSummary) error {
	// Generate an hourly identifier (e.g., "03" for 3:00 AM)
	hourIdentifier := time.Now().Format("1504")
	// Create folder structure: basePath/YYYY-MM-DD/appID/
	dateFolder := time.Now().Format(resultTimeFormat)
	folderPath := filepath.Join(basePath, dateFolder, appID)

	// Ensure directories exist
	if err := os.MkdirAll(folderPath, 0750); err != nil {
		return fmt.Errorf("failed to create folder %s: %w", folderPath, err)
	}

	// Create file: hourIdentifier.json
	fileName := fmt.Sprintf("%s-%s.json", hourIdentifier, familyType)

	// Clean and validate the path
	filePath := filepath.Join(folderPath, filepath.Clean(fileName))

	filePath = filepath.Clean(filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			mng.Logger.Error(err, "Failed to close file", mng.Pack, "saveFile")
		}
	}(file)

	// Serialize data to JSON and write to the file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	fmt.Printf("File created: %s\n", filePath)
	return nil
}

func (mng *LocalResultManager) findAndScrambleAttribute(payload validation.DynamicStruct) validation.DynamicStruct {
	for k, v := range payload {
		if mng.cm.ConfigurationSettings.SecuritySettings.HaveToMask(k) {
			payload[k] = mng.scrambleValue(v) // Scramble the value
			continue
		}

		switch val := v.(type) {
		case map[string]interface{}:
			// Recurse into nested map
			mng.findAndScrambleAttribute(val)
		case []interface{}:
			// Iterate over arrays of objects
			for _, item := range val {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					mng.findAndScrambleAttribute(nestedMap)
				}
			}
		}
	}

	return payload
}

func (mng *LocalResultManager) scrambleValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Mask string (e.g., replace all but the first and last character with *)
		if len(v) > 2 {
			return string(v[0]) + strings.Repeat("*", len(v)-2) + string(v[len(v)-1])
		}

		return strings.Repeat("*", len(v))
	case int, float64:
		// Mask numeric values (e.g., replace with a placeholder or zero out)
		return 0
	case bool:
		// Replace boolean values with false
		return false
	default:
		// Handle other types generically
		return "**********"
	}
}
