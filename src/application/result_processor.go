package application

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/services"
)

// MessageResult contains the information for a validation
type MessageResult struct {
	TransmitterID      string              // Organisation ID of the transmitter
	Endpoint           string              // Name of the endpoint
	HTTPMethod         string              // Type of HTTP method
	Result             bool                // Indicates the result of the validation (True= Valid  ok)
	ServerID           string              // Identifies the server requesting the information
	Errors             map[string][]string // Details for the errors found during the validation
	XFapiInteractionID string
}

// EndpointSummary contains the summary information for the validations by endpoint
type EndpointSummary struct {
	Endpoint       string // Name of the endpoint
	TotalResults   int    // Total results for this specific endpoint
	ValidResults   int    // Total number of validation marked as "true"
	InvalidResults int    // Total number of validation marked as "false"
}

// ErrorDetail Contains de detail for a specific error
type ErrorDetail struct {
	Field     string // Name of the field with problems
	ErrorType string // Description of the error found
}

// TransmitterResults Stores the result for a specific transmitterID
type TransmitterResults struct {
	TransmitterID  string
	GroupedResults map[string][]MessageResult // slice to store grouped results
}

var (
	resultProcessorSingleton ResultProcessor // Singleton instance of the ResultProcessor
	resultProcessorMutex     = sync.Mutex{}  // Mutex for thread-safe access to messageResults
	txGroupedResults         = make(map[string]TransmitterResults)
	totalResults             = 0 // total Number of results validated
)

// ResultProcessor struct in charge of processing results
type ResultProcessor struct {
	crosscutting.OFBStruct
	reportStartTime time.Time             // Datetime of the start of the report
	mqdServer       services.ReportServer // Report server for MQD
	cm              *ConfigurationManager // Manager for application settings
}

// GetResultProcessor returns the singleton instance of the ResultProcessor
//
// Parameters:
//   - logger: Logger to be used by the processor
//   - mqdServer: MQD Server to send the results
//   - cm: Configuration manager
//
// Returns:
//   - *ResultProcessor: New result processor created
func GetResultProcessor(logger log.Logger, mqdServer services.ReportServer, cm *ConfigurationManager) *ResultProcessor {
	if resultProcessorSingleton.Pack == "" {
		resultProcessorSingleton = ResultProcessor{
			OFBStruct: crosscutting.OFBStruct{
				Pack:   "ResultProcessor",
				Logger: logger,
			},
			cm:              cm,
			mqdServer:       mqdServer,
			reportStartTime: time.Time{},
		}
	}

	return &resultProcessorSingleton
}

// AppendResult is for appending a message result
//
// Parameters:
//   - result: Message result to be included
//
// Returns:
func (rp *ResultProcessor) AppendResult(result *MessageResult) {
	resultProcessorMutex.Lock()
	totalResults++

	transmitterID := result.TransmitterID
	if transmitterID == "" {
		transmitterID = rp.cm.settings.ApplicationSettings.OrganisationID
	}

	if _, ok := txGroupedResults[transmitterID]; !ok {
		txGroupedResults[transmitterID] = TransmitterResults{
			TransmitterID:  transmitterID,
			GroupedResults: make(map[string][]MessageResult),
		}
	}

	txResult := txGroupedResults[transmitterID]
	txResult.GroupedResults[result.ServerID] = append(txResult.GroupedResults[result.ServerID], *result)
	txGroupedResults[transmitterID] = txResult

	rp.Logger.Debug("Total grouped Results for TransmitterID: ["+transmitterID+"] in ServerID ["+result.ServerID+"] :"+strconv.Itoa(len(txResult.GroupedResults[result.ServerID])), rp.Pack, "getAndClearResults")
	resultProcessorMutex.Unlock()
}

// GetAndClearResults returns the actual results, and cleans the lists
//
// Parameters:
//
// Returns:
//   - map: map[string][]MessageResult List of message results by clientID
func (rp *ResultProcessor) getAndClearResults() map[string]TransmitterResults {
	rp.Logger.Info("Loading results", rp.Pack, "getAndClearResults")
	resultProcessorMutex.Lock()
	rp.Logger.Debug("Total Results Found :"+strconv.Itoa(totalResults), rp.Pack, "getAndClearResults")
	defer func() {
		//groupedResults = make(map[string][]MessageResult)
		txGroupedResults = make(map[string]TransmitterResults)
		totalResults = 0
		resultProcessorMutex.Unlock()
	}()

	return txGroupedResults
}

// StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
//
// Parameters:
//
// Returns:
func (rp *ResultProcessor) StartResultsProcessor() {
	rp.Logger.Info("Starting result processor, ReportExecutionWindow: "+strconv.Itoa(rp.cm.ConfigurationSettings.ReportSettings.ReportExecutionWindow), rp.Pack, "StartResultsProcessor")
	rp.reportStartTime = time.Now()
	timeWindow := time.Duration(rp.cm.GetReportExecutionWindow()) * time.Minute
	// create an empty result for the initial run
	newResult := TransmitterResults{
		TransmitterID: rp.cm.settings.ApplicationSettings.OrganisationID,
	}

	txGroupedResults[rp.cm.settings.ApplicationSettings.OrganisationID] = newResult
	// Send an initial report for observability.
	rp.processAndSendResults()
	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			rp.processAndSendResults()
		case <-time.After(5 * time.Second):
			if totalResults >= rp.cm.GetSendOnReportNumber() {
				rp.processAndSendResults()
				ticker.Stop()                       // Stop the current ticker
				ticker = time.NewTicker(timeWindow) // Restart the ticker
			}
		}
	}
}

// processAndSendResults Processes the current results (creates a summary report) and sends it to the main server
//
// Parameters:
//
// Returns:
func (rp *ResultProcessor) processAndSendResults() {
	rp.Logger.Info("Processing and sending results", "result", "processAndSendResults")
	processStartTime := time.Now()
	report := models.Report{DataOwnerID: rp.cm.settings.ApplicationSettings.OrganisationID}
	rp.updateMetrics(&report)
	rp.reportStartTime = time.Now()
	results := rp.getAndClearResults()
	rp.Logger.Debug("Total Results to process :"+strconv.Itoa(len(results)), rp.Pack, "processAndSendResults")

	for _, transmitterResult := range results {
		report.ClientID = transmitterResult.TransmitterID
		report.ServerSummary = rp.getSummary(transmitterResult.GroupedResults)
		rp.Logger.Debug("Total ServerSummary process :"+strconv.Itoa(len(report.ServerSummary)), rp.Pack, "processAndSendResults")
		report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.ReportGenerationTime", Value: time.Since(processStartTime).String()})
		err := rp.mqdServer.SendReport(report)
		if err != nil {
			rp.Logger.Error(err, "Error sending report", rp.Pack, "processAndSendResults")
			return
		}
		rp.printReport(report)
	}

	rp.Logger.Info("processAndSendResults -> Process finished", "server", "postReport")
}

// updateMetrics Updates the metrics for the report
//
// Parameters:
//   - report: Report with the metric information
//
// Returns:
func (rp *ResultProcessor) updateMetrics(report *models.Report) {
	rp.Logger.Info("Updating metrics", rp.Pack, "updateMetrics")
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.ReportStartDate", Value: rp.reportStartTime.String()})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.ReportEndDate", Value: time.Now().String()})
	systemMetrics := monitoring.GetAndCleanSystemMetrics()
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.BadRequestErrors", Value: systemMetrics.BadRequestsReceived})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.TotalRequests", Value: systemMetrics.RequestsReceived})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.MemoryUsageAvg", Value: systemMetrics.AverageMemory})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.MemoryUsageMax", Value: systemMetrics.MaxUsedMemory})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.CPUNumber", Value: systemMetrics.AllowedCPUs})
	report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{Key: "runtime.ResponseTimeAvg", Value: systemMetrics.AverageResponseTime})

	report.ApplicationConfiguration.ApplicationVersion = monitoring.Version
	report.ApplicationConfiguration.Environment = rp.cm.settings.ConfigurationSettings.Environment
	report.ApplicationConfiguration.ApplicationID = rp.cm.settings.ConfigurationSettings.ApplicationID.String()
	report.ApplicationConfiguration.ReportExecutionWindow = strconv.Itoa(rp.cm.GetReportExecutionWindow())
	report.ApplicationConfiguration.ReportExecutionNumber = strconv.Itoa(rp.cm.GetSendOnReportNumber())

	report.ApplicationConfiguration.ResultSettingsEnabled = rp.cm.settings.ResultSettings.Enabled
	report.ApplicationConfiguration.ResultSettingsDaysToStore = rp.cm.settings.ResultSettings.DaysToStore
	report.ApplicationConfiguration.ResultSettingsFilesPerDay = rp.cm.settings.ResultSettings.FilesPerDay
	report.ApplicationConfiguration.ResultSettingsSamplesPerError = rp.cm.settings.ResultSettings.SamplesPerError
	report.ApplicationConfiguration.ResultSettingsMaskPrivateContent = rp.cm.settings.ResultSettings.MaskPrivateContent

	report.ApplicationConfiguration.ConfigurationUpdateStatus.LastExecutionDate = rp.cm.GetLastExecutionDate()
	report.ApplicationConfiguration.ConfigurationUpdateStatus.LastUpdatedDate = rp.cm.GetLastUpdatedDate()
	for key, value := range rp.cm.GetUpdateMessages() {
		report.ApplicationConfiguration.ConfigurationUpdateStatus.ConfigurationUpdateError = append(report.ApplicationConfiguration.ConfigurationUpdateStatus.ConfigurationUpdateError, models.ConfigurationUpdateError{
			ErrorDate:    key,
			ErrorMessage: value,
		})
	}

	report.ApplicationConfiguration.ConfigurationUpdateStatus.ConfigurationVersion = rp.cm.ConfigurationSettings.Version
	report.ApplicationConfiguration.ApplicationMode = rp.cm.settings.ApplicationSettings.Mode

	ue := monitoring.GetAndCleanUnsupportedEndpoints()
	for key, date := range ue {
		for versionKey, value := range date {
			errorMessage := "Endpoint not supported"
			if versionKey != "N.A." {
				errorMessage = "Version not supported"
			}
			report.UnsupportedEndpoints = append(report.UnsupportedEndpoints, models.UnsupportedEndpoint{
				EndpointName: key,
				Count:        value,
				Version:      versionKey,
				Error:        errorMessage,
			})
		}
	}
}

// getSummary Returns the server summary for a specific set of MessageResults
//
// Parameters:
//   - results: List of results for a specific server
//
// Returns:
//   - ServerSummary: Summary by each point for the specified server
func (rp *ResultProcessor) getSummary(results map[string][]MessageResult) []models.ServerSummary {
	result := make([]models.ServerSummary, 0)
	for key, messageResult := range results {
		newSummary := models.ServerSummary{ServerID: key}
		for _, endpointResult := range messageResult {
			newSummary.TotalRequests++
			newSummary.EndpointSummary = rp.updateEndpointSummary(newSummary.EndpointSummary, endpointResult)
		}

		result = append(result, newSummary)
	}

	return result
}

// updateEndpointSummary Updates the summary for a specific endpoint
//
// Parameters:
//   - endpointSummary: summary to be updated
//   - messageResult: Result to be included on the summary
//
// Returns:
//   - ServerSummary: Summary updated with the result
func (rp *ResultProcessor) updateEndpointSummary(endpointSummary []models.EndPointSummary, messageResult MessageResult) []models.EndPointSummary {
	newEPSummary := models.EndPointSummary{EndpointName: messageResult.Endpoint, TotalRequests: 1}
	found := false
	for i, ep := range endpointSummary {
		if ep.EndpointName == newEPSummary.EndpointName {
			found = true
			endpointSummary[i].TotalRequests++
			if !messageResult.Result {
				endpointSummary[i].ValidationErrors++
				endpointSummary[i].Detail = rp.updateEndpointSummaryDetail(endpointSummary[i].Detail, messageResult.Errors, messageResult.XFapiInteractionID)
			}

			break
		}
	}

	if !found {
		if !messageResult.Result {
			newEPSummary.ValidationErrors = 1
			newEPSummary.Detail = rp.updateEndpointSummaryDetail(newEPSummary.Detail, messageResult.Errors, messageResult.XFapiInteractionID)
		}

		endpointSummary = append(endpointSummary, newEPSummary)
	}

	return endpointSummary
}

// updateEndpointSummaryDetail Updates the summary detail for a specific endpoint / field
//
// Parameters:
//   - details: Details to be updated
//   - errors: List of errors to be included
//   - xfapiID: xFapi ID of the transaction
//
// Returns:
//   - EndPointSummaryDetail: Updated detail with the errors
func (rp *ResultProcessor) updateEndpointSummaryDetail(details []models.EndPointSummaryDetail, errors map[string][]string, xfapiID string) []models.EndPointSummaryDetail {
	for key, val := range errors {
		newDetail := &models.EndPointSummaryDetail{Field: key}
		fieldFound := false
		for i, field := range details {
			if key == field.Field {
				fieldFound = true
				newDetail = &details[i]
				break
			}
		}

		newDetail.Details = rp.updateFieldDetails(newDetail.Details, val, xfapiID)
		if !fieldFound {
			details = append(details, *newDetail)
		}
	}

	return details
}

// updateFieldDetails Updates the summary detail for a specific field
//
// Parameters:
//   - details: Details to be updated
//   - fieldDetails: Field details to include
//   - xfapiID: xFapi ID of the transaction
//
// Returns:
//   - FieldDetail: Updated FieldDetail with the errors
func (rp *ResultProcessor) updateFieldDetails(details []models.FieldDetail, fieldDetails []string, xfapiID string) []models.FieldDetail {
	for _, errorDetail := range fieldDetails {
		detailFound := false
		for j, fieldDetail := range details {
			if fieldDetail.ErrorType == errorDetail {
				detailFound = true
				details[j].XFapiList = append(details[j].XFapiList, xfapiID)
				details[j].TotalCount++
			}
		}

		if !detailFound {
			details = append(details, models.FieldDetail{ErrorType: errorDetail, TotalCount: 1, XFapiList: []string{xfapiID}})
		}
	}

	return details
}

// printReport Prits the report to console (Should be used for DEBUG pourpuses only)
//
// Parameters:
//   - report: Report to be printed
//
// Returns:
func (rp *ResultProcessor) printReport(report models.Report) {
	b, err := json.Marshal(report)
	if err != nil {
		rp.Logger.Error(err, "Error while printing the report.", rp.Pack, "printReport")
		return
	}

	rp.Logger.Debug(string(b), rp.Pack, "printReport")
}
