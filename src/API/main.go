// api/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/worker"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

/**
 * Func: handleMessages Handles requests to the specified urls in the settings
 *
 * @author AB
 *
 * @params
 * w: Writer to create the response
 * r: Request received
 * @return
 */
func handleMessages(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	monitoring.IncreaseRequestsReceived()
	var msg queue.Message

	clientOrgId := r.Header.Get("clientOrgId")
	_, err := uuid.Parse(clientOrgId)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		fmt.Println("clientOrgId: " + clientOrgId)
		fmt.Println("Error: clientOrgId: Not found or bad format.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("clientOrgId: Not found or bad format."))
		return
	}

	routeName := mux.CurrentRoute(r).GetName()

	jsonData, err := buildJSONMsg(r)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	msg.HeaderMessage = jsonData
	msg.Endpoint = routeName
	msg.HTTPMethod = r.Method
	msg.ClientID = clientOrgId

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)
	monitoring.RecordResponseDuration(startTime)
	fmt.Fprintf(w, "Message enqueued for processing!")
}

func buildJSONMsg(request *http.Request) (string, error) {
	// Create a map to store header key-value pairs.
	objectMap := make(map[string]interface{})
	headermap := getHeaderMap(request.Header)
	for k, v := range headermap {
		objectMap[k] = v
	}
	pathMap := getPathMap(mux.Vars(request))
	for k, v := range pathMap {
		objectMap[k] = v
	}

	// Convert the map to a JSON string.
	jsonData, err := json.Marshal(objectMap)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

/**
* Func: headersToJSON Maps all the headers found in the request to a JSON object
*
* @author AB
*
* @params
* headers: List of headers to map
* @return
* string: JSON object created
* error: Returs error in case a problem is found during the mapping
 */
func getHeaderMap(headers http.Header) map[string]interface{} {
	// Create a map to store header key-value pairs.
	headerMap := make(map[string]interface{})

	// Iterate through the header parameters and add them to the map.
	for key, values := range headers {

		key = strings.ToLower(key)
		// If there's only one value for the header, store it directly.
		if len(values) == 1 {
			headerMap[key] = values[0]
		} else {
			// If there are multiple values, store them as an array.
			headerMap[key] = values
		}
	}

	return headerMap
}

/**
* Func: headersToJSON Maps all the headers found in the request to a JSON object
*
* @author AB
*
* @params
* headers: List of headers to map
* @return
* string: JSON object created
* error: Returs error in case a problem is found during the mapping
 */
func getPathMap(values map[string]string) map[string]interface{} {
	// Create a map to store header key-value pairs.
	pathMap := make(map[string]interface{})

	// Iterate through the header parameters and add them to the map.
	for key, values := range values {

		key = strings.ToLower(key)
		// If there's only one value for the header, store it directly.
		if len(values) == 1 {
			pathMap[key] = values[0]
		} else {
			// If there are multiple values, store them as an array.
			pathMap[key] = values
		}
	}

	return pathMap
}

/**
* Func: Main is the main function of the api, that is executed on "run"
*
* @author AB
*
* @params
 * @return
*/
func main() {
	monitoring.StartOpenTelemetry()

	configuration.Initialize()
	// Start the worker Goroutine to process messages
	go worker.StartWorker()
	go result.StartResultsProcessor()

	r := mux.NewRouter()
	//http.Handle("/metrics", promhttp.Handler())
	r.Handle("/metrics", monitoring.GetOpentelemetryHandler())
	r.HandleFunc("/sendmessage", handleMessages).Name("/sendmessage").Methods("GET")
	for _, element := range configuration.GetEndpointSettings() {
		r.HandleFunc(element.Endpoint, handleMessages).Name(element.Endpoint).Methods("GET")
		println("handling endpoint: " + element.Endpoint)
	}

	//http.Handle("/", r)
	port := crosscutting.GetEnvironmentValue("API_PORT", ":8080")

	fmt.Println("Starting the server on port " + port)
	log.Fatal(http.ListenAndServe(port, r))
}
