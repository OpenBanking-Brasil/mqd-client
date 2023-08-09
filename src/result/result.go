package result

import (
	"fmt"
	"sync"
	"time"
)

// MessageResult contains the information for a validation
type MessageResult struct {
	Endpoint   string // Name of the endpoint
	HTTPMethod string // Type of HTTP method
	Result     bool   // Indicates the result of the validation (True= Valid  ok)
	ClientID   string // Identifier of the Client requesting the information
	ServerID   string // Identifies the server requesting the information
}

// EndpointSummary contains the summary information for the validations by endpoint
type EndpointSummary struct {
	Endpoint       string // Name of the endpoint
	TotalResults   int    // Total results for this specific endpoint
	ValidResults   int    // Total number of validation marked as "true"
	InvalidResults int    // Total number of validation marked as "false"
}

var (
	messageResults = make([]MessageResult, 0) // Slice to store message results
	mutex          = sync.Mutex{}             // Mutex for thread-safe access to messageResults
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
func GetAndClearResults() []MessageResult {
	mutex.Lock()
	defer func() {
		messageResults = nil // Clear the results
		mutex.Unlock()
	}()

	return messageResults
}

/**
 * Func: StartResultsProcessor starts the periodic process that prints total results and clears them every 2 minutes
 *
 * @author AB
 */
func StartResultsProcessor() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			results := GetAndClearResults()
			printTotalResults(results)
		}
	}
}

/**
 * Func: printTotalResults is for printing the summary of the results
 *
 * @author GoCommnets
 *
 * @params
 * results: Lists of results to be printed
 * @return
 */
func printTotalResults(results []MessageResult) {
	fmt.Println("Results:.")
	summaryMap := make(map[string]EndpointSummary)

	for _, result := range results {
		summary, ok := summaryMap[result.Endpoint]
		if !ok {
			summary.Endpoint = result.Endpoint
		}
		summary.TotalResults++
		if result.Result {
			summary.ValidResults++
		} else {
			summary.InvalidResults++
		}
		summaryMap[result.Endpoint] = summary
	}

	// fmt.Println("Results Summary:")
	for _, summary := range summaryMap {
		fmt.Printf(
			"Endpoint: %s, Total: %d, Valid: %d, Invalid: %d\n",
			summary.Endpoint, summary.TotalResults, summary.ValidResults, summary.InvalidResults,
		)
	}
}
