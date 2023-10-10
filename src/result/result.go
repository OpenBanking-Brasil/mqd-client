package result

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
)

// MessageResult contains the information for a validation
type MessageResult struct {
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

// Contains the name and value for different types of metrics for the report
type MetricObject struct {
	Key   string // Name of the metric
	Value string // Value of the metric
}

// Contains a list of metrics recorded for the report
type ApplicationMetrics struct {
	Values []MetricObject // List of metrics with its values
}

// Contains de detail for a specific error
type ErrorDetail struct {
	Field     string // Name of the field with problems
	ErrorType string // Description of the error found
}

// Detail for a filed with an error type
type FieldDetail struct {
	ErrorType  string   // Name of the error type found
	XFapiList  []string // List of xFapiInteractionIds that showed this specific error
	TotalCount int      // Number of times the error was found
}

// Summary of the details of errors for a specific field
type EndPointSummaryDetail struct {
	Field   string        // Name of the field
	Details []FieldDetail // List of details with the errors
}

// Contains a summary for a specific endpoint
type EndPointSummary struct {
	EndpointName     string                  // Name of the endpoint
	TotalRequests    int                     // Totla number of requests
	ValidationErrors int                     // Total number of validation errors
	Detail           []EndPointSummaryDetail // Detail of the errors
}

// Summary of a specific server
type ServerSummary struct {
	ServerId        string            // Server identifier (UUID)
	TotalRequests   int               // Total number of requests
	EndpointSummary []EndPointSummary // Summary of the endpoints requested
}

// Object report to be sent to the server
type Report struct {
	Metrics       ApplicationMetrics // Metris of the application
	ClientID      string             // Client identifier (UUID)
	ServerSummary []ServerSummary    // List of Servers requested
}

var (
	messageResults  = make([]MessageResult, 0)         // Slice to store message results
	mutex           = sync.Mutex{}                     // Mutex for thread-safe access to messageResults
	groupedResults  = make(map[string][]MessageResult) // slice to store grouped results
	reportStartTime = time.Time{}                      // Datetime of the start of the report
)

// Func: AppendResult is for appending a message result
// @author AB
// @params
// result: Message rresult to be included
// @return
func AppendResult(result *MessageResult) {
	mutex.Lock()
	messageResults = append(messageResults, *result)
	groupedResults[result.ServerID] = append(groupedResults[result.ServerID], *result)
	mutex.Unlock()
}

// Func: GetAndClearResults returns the actual results, and cleans the lists
// @author AB
// @return
// List of Message results
func getAndClearResults() map[string][]MessageResult {
	mutex.Lock()
	defer func() {
		messageResults = nil // Clear the results
		groupedResults = make(map[string][]MessageResult)
		mutex.Unlock()
	}()

	return groupedResults
}

// Func: StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
// @author AB
func StartResultsProcessor() {
	loadCertificates()
	reportStartTime = time.Now()
	timeWindow := time.Duration(configuration.ReportExecutiontimeFrame) * time.Minute
	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			processAndSendResults()
		}
	}
}

// Func: processAndSendResults Processes the current results (creates a summary report) and sends it to the main server
// @author AB
// @params
// @return
func processAndSendResults() {
	log.Info("Processing and sending results", "result", "processAndSendResults")
	processStartTime := time.Now()
	report := Report{ClientID: configuration.ClientID}
	updateMetrics(&report.Metrics)
	reportStartTime = time.Now()
	results := getAndClearResults()
	report.ServerSummary = getSummary(results)
	report.Metrics.Values = append(report.Metrics.Values, MetricObject{Key: "ReportGenerationtime", Value: time.Since(processStartTime).String()})

	sendReportToAPI(report)
	printReport(report)
	log.Info("processAndSendResults -> Process finished", "server", "postReport")
}

// Func: updateMetrics Updates the metrics for the report
// @author AB
// @params
// metrics: List of metrics to be included
// @return
func updateMetrics(metrics *ApplicationMetrics) {
	metrics.Values = append(metrics.Values, MetricObject{Key: "ReportStartDate", Value: reportStartTime.String()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "ReportEndDate", Value: time.Now().String()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "BadRequestErrors", Value: strconv.Itoa(monitoring.GetAndCleanBadRequestsReceived())})
	metrics.Values = append(metrics.Values, MetricObject{Key: "TotalRequests", Value: strconv.Itoa(monitoring.GetAndCleanRequestsReceived())})
	metrics.Values = append(metrics.Values, MetricObject{Key: "MemmoryUsageAvg", Value: monitoring.GetAndCleanAverageMemmory()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "ResposeTimeAvg", Value: monitoring.GetAndCleanResponseTime()})
}

// Func: getSummary Returns the server summary for a specific set of MessageResults
// @author AB
// @params
// results: List of results for a specific server
// @return
// ServerSummary: Summary by each point for the speciified server
func getSummary(results map[string][]MessageResult) []ServerSummary {
	result := make([]ServerSummary, 0)
	for key, messageResult := range results {
		newSummary := ServerSummary{ServerId: key}
		for _, endpointResult := range messageResult {
			newSummary.TotalRequests++
			newSummary.EndpointSummary = updateEndpointSummary(newSummary.EndpointSummary, endpointResult)
		}

		result = append(result, newSummary)
	}

	return result
}

// Func: updateEndpointSummary Updates the summary for a specific endpoint
// @author AB
// @params
// endpointSummary: summary to be updated
// messageResult: Result to be included on the summary
// @return
// ServerSummary: Summary updated with the result
func updateEndpointSummary(endpointSummary []EndPointSummary, messageResult MessageResult) []EndPointSummary {
	newepSummary := EndPointSummary{EndpointName: messageResult.Endpoint, TotalRequests: 1}
	found := false
	for i, ep := range endpointSummary {
		if ep.EndpointName == newepSummary.EndpointName {
			found = true
			endpointSummary[i].TotalRequests = endpointSummary[i].TotalRequests + 1
			if !messageResult.Result {
				endpointSummary[i].ValidationErrors = endpointSummary[i].ValidationErrors + 1
				endpointSummary[i].Detail = updateEndpointSummaryDetail(endpointSummary[i].Detail, messageResult.Errors, messageResult.XFapiInteractionID)
			}

			break
		}
	}

	if !found {
		if !messageResult.Result {
			newepSummary.ValidationErrors = 1
			newepSummary.Detail = updateEndpointSummaryDetail(newepSummary.Detail, messageResult.Errors, messageResult.XFapiInteractionID)
		}

		endpointSummary = append(endpointSummary, newepSummary)
	}

	return endpointSummary
}

// Func: updateEndpointSummaryDetail Updates the summary detail for a specific endpoint / field
// @author AB
// @params
// details: Details to be updated
// errors: List of errors to be included
// @return
// EndPointSummaryDetail: Updated detail with the errors
func updateEndpointSummaryDetail(details []EndPointSummaryDetail, errors map[string][]string, xfapiID string) []EndPointSummaryDetail {
	for key, val := range errors {
		newDetail := &EndPointSummaryDetail{Field: key}
		fieldFound := false
		for i, field := range details {
			if key == field.Field {
				fieldFound = true
				newDetail = &details[i]
				break
			}
		}

		newDetail.Details = updateFieldDetails(newDetail.Details, val, xfapiID)
		if !fieldFound {
			details = append(details, *newDetail)
		}
	}

	return details
}

// Func: updateFieldDetails Updates the summary detail for a specific field
// @author AB
// @params
// details: Field Details to be updated
// fieldDetails: Field details to include
// @return
// FieldDetail: Updated FieldDetail with the errors
func updateFieldDetails(details []FieldDetail, fieldDetails []string, xfapiID string) []FieldDetail {
	for _, errorDetail := range fieldDetails {
		detailFound := false
		for j, fieldDetail := range details {
			if fieldDetail.ErrorType == errorDetail {
				detailFound = true
				details[j].XFapiList = append(details[j].XFapiList, xfapiID)
				details[j].TotalCount = details[j].TotalCount + 1
			}
		}

		if !detailFound {
			details = append(details, FieldDetail{ErrorType: errorDetail, TotalCount: 1, XFapiList: []string{xfapiID}})
		}
	}

	return details
}

// Func: printReport Prits the report to console (Should be used for DEBUG pourpuses only)
// @author AB
// @params
// report: Report to be printed
// @return
func printReport(report Report) {
	b, err := json.Marshal(report)
	if err != nil {
		log.Error(err, "Error while printing the report.", "Result", "printReport")
		return
	}

	log.Debug(string(b), "Result", "printReport")
}
