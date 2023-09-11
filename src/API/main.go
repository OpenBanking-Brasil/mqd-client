// api/main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
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
 * Func: handleValidateResponseMessage Handles requests to the specified urls in the settings
 *
 * @author AB
 *
 * @params
 * w: Writer to create the response
 * r: Request received
 * @return
 */
func handleValidateResponseMessage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	monitoring.IncreaseRequestsReceived()
	var msg queue.Message

	// Read the Server Organization ID from the header
	serverOrgId := r.Header.Get("serverOrgId")
	_, err := uuid.Parse(serverOrgId)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("serverOrgId: Not found or bad format."))
		return
	}

	// Read the Server Organization ID from the header
	endpointName := r.Header.Get("endpointName")

	// Validate the endpoint configuration exists
	endpointSettings := configuration.GetEndpointSetting(endpointName)

	if endpointSettings.Endpoint == "" {
		monitoring.IncreaseBadRequestsReceived()
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("endpointName: Not found or bad format."))
		return
	}

	// Read header and create a json object
	headerMsg, err := buildHeaderMsg(&r.Header)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read the body of the message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	msg.HeaderMessage = headerMsg
	msg.Message = string(body)
	msg.Endpoint = endpointName
	msg.HTTPMethod = r.Method
	msg.ServerID = serverOrgId

	// // print the message for validation
	// b, err := json.Marshal(msg)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(b))

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)
	monitoring.RecordResponseDuration(startTime)
	fmt.Fprintf(w, "Message enqueued for processing!")
}

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

	// Get Route name, as the EndPoint Name
	routeName := mux.CurrentRoute(r).GetName()

	// Read the Server Organization ID from the header
	serverOrgId := r.Header.Get("serverOrgId")
	_, err := uuid.Parse(serverOrgId)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("serverOrgId: Not found or bad format."))
		return
	}

	// Read header and create a json object
	headerMsg, err := buildHeaderMsg(&r.Header)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read the body of the message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	msg.HeaderMessage = headerMsg
	msg.Message = string(body)
	msg.Endpoint = routeName
	msg.HTTPMethod = r.Method
	// msg.ClientID = serverOrgId

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)
	monitoring.RecordResponseDuration(startTime)
	fmt.Fprintf(w, "Message enqueued for processing!")
}

func buildHeaderMsg(header *http.Header) (string, error) {
	// Create a map to store header key-value pairs.
	objectMap := make(map[string]interface{})
	headermap := getHeaderMap(*header)
	for k, v := range headermap {
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
// func getPathMap(values map[string]string) map[string]interface{} {
// 	// Create a map to store header key-value pairs.
// 	pathMap := make(map[string]interface{})

// 	// Iterate through the header parameters and add them to the map.
// 	for key, values := range values {

// 		key = strings.ToLower(key)
// 		// If there's only one value for the header, store it directly.
// 		if len(values) == 1 {
// 			pathMap[key] = values[0]
// 		} else {
// 			// If there are multiple values, store them as an array.
// 			pathMap[key] = values
// 		}
// 	}

// 	return pathMap
// }

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

	// Validator for Responses
	r.HandleFunc("/ValidateResponse", handleValidateResponseMessage)

	for _, element := range configuration.GetEndpointSettings() {
		r.HandleFunc(element.Endpoint, handleMessages).Name(element.Endpoint).Methods("GET")
		println("handling endpoint: " + element.Endpoint)
	}

	//http.Handle("/", r)
	port := crosscutting.GetEnvironmentValue("API_PORT", ":8080")

	fmt.Println("Starting the server on port " + port)
	log.Fatal(http.ListenAndServe(port, r))
}
