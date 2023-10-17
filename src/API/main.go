// api/main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/worker"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Contains information message when error needs to be returned
type GenericError struct {
	Message string // Error message
}

func updateReponseError(w http.ResponseWriter, genericError GenericError, responseCode int) {
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(genericError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code
	w.WriteHeader(responseCode)

	// Write the JSON data to the response
	w.Write(jsonData)
}

// Func: handleValidateResponseMessage Handles requests to the specified urls in the settings
// @author AB
// @params
// w: Writer to create the response
// r: Request received
// @return
func handleValidateResponseMessage(w http.ResponseWriter, r *http.Request) {
	genericError := &GenericError{}
	startTime := time.Now()
	monitoring.IncreaseRequestsReceived()
	var msg queue.Message

	// Read the Server Organization ID from the header
	serverOrgId := r.Header.Get("serverOrgId")
	_, err := uuid.Parse(serverOrgId)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "serverOrgId: Not found or bad format."
		updateReponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read the Server Organization ID from the header
	endpointName := r.Header.Get("endpointName")

	// Validate the endpoint configuration exists
	endpointSettings := configuration.GetEndpointSetting(endpointName)

	if endpointSettings.Endpoint == "" {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "endpointName: Not found or bad format."
		updateReponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read header and create a json object
	headerMsg, err := buildHeaderMsg(&r.Header)
	if err != nil {
		log.Error(err, "Error: handleValidateResponseMessage", "API", "handleValidateResponseMessage")
		genericError.Message = "Error processing request."
		updateReponseError(w, *genericError, http.StatusInternalServerError)
		return
	}

	// Read the body of the message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		genericError.Message = "Failed to read request body."
		updateReponseError(w, *genericError, http.StatusInternalServerError)
		return
	}

	// Read the Server Organization ID from the header
	msg.HeaderMessage = headerMsg
	msg.Message = string(body)
	msg.Endpoint = endpointName
	msg.HTTPMethod = r.Method
	msg.ServerID = serverOrgId
	xFapiID := r.Header.Get("x-fapi-interaction-id")
	// log.Info(xFapiID, "API", "Main")
	msg.XFapiInteractionID = xFapiID

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)
	monitoring.RecordResponseDuration(startTime)
	fmt.Fprintf(w, "Message enqueued for processing!")
}

// Func: handleMessages Handles requests to the specified urls in the settings
// @author AB
// @params
// w: Writer to create the response
// r: Request received
// @return
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
		log.Error(err, "Error: handleMessages", "API", "handleMessages")
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

// Func: buildHeaderMsg Creates a JSON message based on the headers
// @author AB
// @params
// header: List of headers
// @return
// string: JSON Message created
// error: in case of parsing error
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

// Func: headersToJSON Maps all the headers found in the request to a JSON object
// @author AB
// @params
// headers: List of headers to map
// @return
// string: JSON object created
// error: Returs error in case a problem is found during the mapping
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

// Func: Main is the main function of the api, that is executed on "run"
// @author AB
// @params
// @return
func main() {
	monitoring.StartOpenTelemetry()

	configuration.Initialize()

	// Start the worker Goroutine to process messages
	go worker.StartWorker()
	go result.StartResultsProcessor()

	r := mux.NewRouter()
	r.Handle("/metrics", monitoring.GetOpentelemetryHandler())

	// Validator for Responses
	r.HandleFunc("/ValidateResponse", handleValidateResponseMessage).Name("ValidateResponse").Methods("POST")

	//// TODO handlers for specific endpoints were removed as /ValidateResponse will validate all requests
	// for _, element := range configuration.GetEndpointSettings() {
	// 	r.HandleFunc(element.Endpoint, handleMessages).Name(element.Endpoint).Methods("POST")
	// 	log.Log("handling endpoint: "+element.Endpoint, "API", "Main")
	// }

	port := crosscutting.GetEnvironmentValue("API_PORT", ":8080")

	log.Log("Starting the server on port "+port, "API", "Main")
	log.Fatal(http.ListenAndServe(port, r), "", "API", "Main")
}
