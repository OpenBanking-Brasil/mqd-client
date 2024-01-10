package result

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/server"
)

const version = "1.1.0"

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

// Contains de detail for a specific error
type ErrorDetail struct {
	Field     string // Name of the field with problems
	ErrorType string // Description of the error found
}

var (
	singleton      ResultProcessor                    // Singleton instance of the ResultProcessor
	mutex          = sync.Mutex{}                     // Mutex for thread-safe access to messageResults
	groupedResults = make(map[string][]MessageResult) // slice to store grouped results
)

// struct in charge of processing results
type ResultProcessor struct {
	pack            string     // name of thes package
	logger          log.Logger // Logger to be used by the processor
	reportStartTime time.Time  // Datetime of the start of the report
}

// GetResultProcessor returns the singleton instance of the ResultProcessor
// @author AB
// @params
// logger: Logger to be used by the processor
// startTime: Initial start time for the request
// @return
// ResultProcessor instance
func GetResultProcessor(logger log.Logger) *ResultProcessor {
	if singleton.pack == "" {
		singleton = ResultProcessor{
			pack:            "ResultProcessor",
			logger:          logger,
			reportStartTime: time.Time{},
		}
	}

	return &singleton
}

// AppendResult is for appending a message result
// @author AB
// @params
// result: Message rresult to be included
// @return
func (rp *ResultProcessor) AppendResult(result *MessageResult) {
	mutex.Lock()
	groupedResults[result.ServerID] = append(groupedResults[result.ServerID], *result)
	rp.logger.Debug("Total grouped Results in serverid ["+result.ServerID+"] :"+strconv.Itoa(len(groupedResults[result.ServerID])), rp.pack, "getAndClearResults")
	mutex.Unlock()
}

// GetAndClearResults returns the actual results, and cleans the lists
// @author AB
// @return
// List of Message results
func (rp *ResultProcessor) getAndClearResults() map[string][]MessageResult {
	rp.logger.Info("Loading results", rp.pack, "getAndClearResults")
	mutex.Lock()
	rp.logger.Debug("Total Results Found :"+strconv.Itoa(len(groupedResults)), rp.pack, "getAndClearResults")
	defer func() {
		groupedResults = make(map[string][]MessageResult)
		mutex.Unlock()
	}()

	return groupedResults
}

// StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
// @author AB
// @params
// @return
func (rp *ResultProcessor) StartResultsProcessor() {
	rp.logger.Info("Starting result processor", rp.pack, "StartResultsProcessor")
	rp.reportStartTime = time.Now()
	timeWindow := time.Duration(configuration.ReportExecutiontimeFrame) * time.Minute

	// Send a initial report for observability.
	rp.processAndSendResults(*server.GetReportServer(rp.logger))
	ticker := time.NewTicker(timeWindow)
	for {
		select {
		case <-ticker.C:
			rp.processAndSendResults(*server.GetReportServer(rp.logger))
		}
	}
}

// processAndSendResults Processes the current results (creates a summary report) and sends it to the main server
// @author AB
// @params
// reportServer: Server to send the report
// @return
func (rp *ResultProcessor) processAndSendResults(reportServer server.ReportServer) {
	rp.logger.Info("Processing and sending results", "result", "processAndSendResults")
	processStartTime := time.Now()
	report := server.Report{ClientID: configuration.ClientID}
	rp.updateMetrics(&report.Metrics)
	rp.reportStartTime = time.Now()
	results := rp.getAndClearResults()
	rp.logger.Debug("Total Results to process :"+strconv.Itoa(len(results)), rp.pack, "processAndSendResults")
	report.ServerSummary = rp.getSummary(results)
	rp.logger.Debug("Total ServerSummary processe :"+strconv.Itoa(len(report.ServerSummary)), rp.pack, "processAndSendResults")
	report.Metrics.Values = append(report.Metrics.Values, server.MetricObject{Key: "runtime.ReportGenerationtime", Value: time.Since(processStartTime).String()})

	reportServer.SendReport(report)
	rp.printReport(report)
	rp.logger.Info("processAndSendResults -> Process finished", "server", "postReport")
}

// updateMetrics Updates the metrics for the report
// @author AB
// @params
// metrics: List of metrics to be included
// @return
func (rp *ResultProcessor) updateMetrics(metrics *server.ApplicationMetrics) {
	rp.logger.Info("Updating metrics", rp.pack, "updateMetrics")
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.ReportStartDate", Value: rp.reportStartTime.String()})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.ReportEndDate", Value: time.Now().String()})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.BadRequestErrors", Value: strconv.Itoa(monitoring.GetAndCleanBadRequestsReceived())})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.TotalRequests", Value: strconv.Itoa(monitoring.GetAndCleanRequestsReceived())})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.MemmoryUsageAvg", Value: monitoring.GetAndCleanAverageMemmory()})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "runtime.ResposeTimeAvg", Value: monitoring.GetAndCleanResponseTime()})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "Configuration.Version", Value: version})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "Configuration.Environment", Value: configuration.Environment})
	metrics.Values = append(metrics.Values, server.MetricObject{Key: "Configuration.REPORT_EXECUTION_WINDOW", Value: strconv.Itoa(configuration.ReportExecutiontimeFrame)})
}

// getSummary Returns the server summary for a specific set of MessageResults
// @author AB
// @params
// results: List of results for a specific server
// @return
// ServerSummary: Summary by each point for the speciified server
func (rp *ResultProcessor) getSummary(results map[string][]MessageResult) []server.ServerSummary {
	result := make([]server.ServerSummary, 0)
	for key, messageResult := range results {
		newSummary := server.ServerSummary{ServerId: key}
		for _, endpointResult := range messageResult {
			newSummary.TotalRequests++
			newSummary.EndpointSummary = rp.updateEndpointSummary(newSummary.EndpointSummary, endpointResult)
		}

		result = append(result, newSummary)
	}

	return result
}

// updateEndpointSummary Updates the summary for a specific endpoint
// @author AB
// @params
// endpointSummary: summary to be updated
// messageResult: Result to be included on the summary
// @return
// ServerSummary: Summary updated with the result
func (rp *ResultProcessor) updateEndpointSummary(endpointSummary []server.EndPointSummary, messageResult MessageResult) []server.EndPointSummary {
	newepSummary := server.EndPointSummary{EndpointName: messageResult.Endpoint, TotalRequests: 1}
	found := false
	for i, ep := range endpointSummary {
		if ep.EndpointName == newepSummary.EndpointName {
			found = true
			endpointSummary[i].TotalRequests = endpointSummary[i].TotalRequests + 1
			if !messageResult.Result {
				endpointSummary[i].ValidationErrors = endpointSummary[i].ValidationErrors + 1
				endpointSummary[i].Detail = rp.updateEndpointSummaryDetail(endpointSummary[i].Detail, messageResult.Errors, messageResult.XFapiInteractionID)
			}

			break
		}
	}

	if !found {
		if !messageResult.Result {
			newepSummary.ValidationErrors = 1
			newepSummary.Detail = rp.updateEndpointSummaryDetail(newepSummary.Detail, messageResult.Errors, messageResult.XFapiInteractionID)
		}

		endpointSummary = append(endpointSummary, newepSummary)
	}

	return endpointSummary
}

// updateEndpointSummaryDetail Updates the summary detail for a specific endpoint / field
// @author AB
// @params
// details: Details to be updated
// errors: List of errors to be included
// @return
// EndPointSummaryDetail: Updated detail with the errors
func (rp *ResultProcessor) updateEndpointSummaryDetail(details []server.EndPointSummaryDetail, errors map[string][]string, xfapiID string) []server.EndPointSummaryDetail {
	for key, val := range errors {
		newDetail := &server.EndPointSummaryDetail{Field: key}
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
// @author AB
// @params
// details: Field Details to be updated
// fieldDetails: Field details to include
// @return
// FieldDetail: Updated FieldDetail with the errors
func (rp *ResultProcessor) updateFieldDetails(details []server.FieldDetail, fieldDetails []string, xfapiID string) []server.FieldDetail {
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
			details = append(details, server.FieldDetail{ErrorType: errorDetail, TotalCount: 1, XFapiList: []string{xfapiID}})
		}
	}

	return details
}

// printReport Prits the report to console (Should be used for DEBUG pourpuses only)
// @author AB
// @params
// report: Report to be printed
// @return
func (rp *ResultProcessor) printReport(report server.Report) {
	b, err := json.Marshal(report)
	if err != nil {
		rp.logger.Error(err, "Error while printing the report.", "Result", "printReport")
		return
	}

	rp.logger.Debug(string(b), "Result", "printReport")
}
