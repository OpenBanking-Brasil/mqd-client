package result

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
)

// MessageResult contains the information for a validation
type MessageResult struct {
	Endpoint   string // Name of the endpoint
	HTTPMethod string // Type of HTTP method
	Result     bool   // Indicates the result of the validation (True= Valid  ok)
	// ClientID   string // Identifier of the Client requesting the information
	ServerID string // Identifies the server requesting the information
	Errors   map[string][]string
}

// EndpointSummary contains the summary information for the validations by endpoint
type EndpointSummary struct {
	Endpoint       string // Name of the endpoint
	TotalResults   int    // Total results for this specific endpoint
	ValidResults   int    // Total number of validation marked as "true"
	InvalidResults int    // Total number of validation marked as "false"
}

type MetricObject struct {
	Key   string
	Value string
}

type ApplicationMetrics struct {
	// ReportStartDate      time.Time
	// ReportEndDate        time.Time
	// MemmoryUsageAvg      string
	// CPUUsageAvg          string
	// TotalRequests        int
	// BadRequestErrors     int
	// ReportGenerationtime string
	// MsgProcessAvg        string
	// ResposeTimeAvg       string
	Values []MetricObject
}

type ErrorDetail struct {
	Field     string
	ErrorType string
}

type FieldDetail struct {
	ErrorType  string
	TotalCount int
}

type EndPointSummaryDetail struct {
	Field   string
	Details []FieldDetail
}

type EndPointSummary struct {
	EndpointName     string
	TotalRequests    int
	ValidationErrors int
	Detail           []EndPointSummaryDetail
}

type ServerSummary struct {
	ServerId        string
	TotalRequests   int
	EndpointSummary []EndPointSummary
}

type Report struct {
	Metrics       ApplicationMetrics
	ClientID      string
	ServerSummary []ServerSummary
}

var (
	messageResults = make([]MessageResult, 0) // Slice to store message results
	mutex          = sync.Mutex{}             // Mutex for thread-safe access to messageResults

	// Create a map to group results by ID
	groupedResults  = make(map[string][]MessageResult)
	reportStartTime = time.Time{}
)

/**
 * Func: AppendResult is for appending a message result
 *
 * @author AB
 *
 * @params
 * result: Message rresult to be included
 * @return
 */
func AppendResult(result *MessageResult) {
	mutex.Lock()
	messageResults = append(messageResults, *result)
	groupedResults[result.ServerID] = append(groupedResults[result.ServerID], *result)
	mutex.Unlock()
}

/**
 * Func: GetAndClearResults returns the actual results, and cleans the lists
 *
 * @author AB
 *
 * @return
 * List of Message results
 */
func getAndClearResults() map[string][]MessageResult {
	mutex.Lock()
	defer func() {
		messageResults = nil // Clear the results
		groupedResults = make(map[string][]MessageResult)
		mutex.Unlock()
	}()

	return groupedResults
}

/**
 * Func: StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
 *
 * @author AB
 */
func StartResultsProcessor() {
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

func processAndSendResults() {
	processStartTime := time.Now()
	report := Report{ClientID: configuration.ClientID}
	updateMetrics(&report.Metrics)
	reportStartTime = time.Now()
	results := getAndClearResults()
	report.ServerSummary = getSummary(results)
	report.Metrics.Values = append(report.Metrics.Values, MetricObject{Key: "ReportGenerationtime", Value: time.Since(processStartTime).String()})

	printReport(report)
}

func updateMetrics(metrics *ApplicationMetrics) {
	metrics.Values = append(metrics.Values, MetricObject{Key: "ReportStartDate", Value: reportStartTime.String()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "ReportEndDate", Value: time.Now().String()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "BadRequestErrors", Value: strconv.Itoa(monitoring.GetAndCleanBadRequestsReceived())})
	metrics.Values = append(metrics.Values, MetricObject{Key: "TotalRequests", Value: strconv.Itoa(monitoring.GetAndCleanRequestsReceived())})
	metrics.Values = append(metrics.Values, MetricObject{Key: "MemmoryUsageAvg", Value: monitoring.GetAndCleanAverageMemmory()})
	metrics.Values = append(metrics.Values, MetricObject{Key: "ResposeTimeAvg", Value: monitoring.GetAndCleanResponseTime()})
}

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

func updateEndpointSummary(endpointSummary []EndPointSummary, messageResult MessageResult) []EndPointSummary {
	newepSummary := EndPointSummary{EndpointName: messageResult.Endpoint, TotalRequests: 1}
	found := false
	for i, ep := range endpointSummary {
		if ep.EndpointName == newepSummary.EndpointName {
			found = true
			endpointSummary[i].TotalRequests = endpointSummary[i].TotalRequests + 1
			if !messageResult.Result {
				endpointSummary[i].ValidationErrors = endpointSummary[i].ValidationErrors + 1
				endpointSummary[i].Detail = updateEndpointSummaryDetal(endpointSummary[i].Detail, messageResult.Errors)
			}

			break
		}
	}

	if !found {
		if !messageResult.Result {
			newepSummary.ValidationErrors = 1
			newepSummary.Detail = updateEndpointSummaryDetal(newepSummary.Detail, messageResult.Errors)
		}

		endpointSummary = append(endpointSummary, newepSummary)
	}

	return endpointSummary
}

func updateEndpointSummaryDetal(details []EndPointSummaryDetail, errors map[string][]string) []EndPointSummaryDetail {
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

		newDetail.Details = updateFieldDetails(newDetail.Details, val)
		if !fieldFound {
			details = append(details, *newDetail)
		}
	}

	return details
}

func updateFieldDetails(details []FieldDetail, fieldDetails []string) []FieldDetail {
	for _, errorDetail := range fieldDetails {
		detailFound := false
		for j, fieldDetail := range details {
			if fieldDetail.ErrorType == errorDetail {
				detailFound = true
				details[j].TotalCount = details[j].TotalCount + 1
			}
		}

		if !detailFound {
			details = append(details, FieldDetail{ErrorType: errorDetail, TotalCount: 1})
		}
	}

	return details
}

func printReport(report Report) {
	b, err := json.Marshal(report)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}
