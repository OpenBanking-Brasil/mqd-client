package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/validation/settings"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Contains information message when error needs to be returned
type GenericError struct {
	Message string // Error message
}

// APIServer Contains the APIServer
type APIServer struct {
	pack           string       // Package name
	logger         log.Logger   // Logger to be used
	metricsHandler http.Handler // Handler for the metric endpoint
	cm             *settings.ConfigurationManager
}

// GetAPIServer Creates a new APIServer
// @author AB
// @param
// logger: Logger to be used
// metricsHandler: Handler for the metric endpoint
// @return
// *APIServer: APIServer
func GetAPIServer(logger log.Logger, metricsHandler http.Handler, cm *settings.ConfigurationManager) *APIServer {
	return &APIServer{
		pack:           "API",
		logger:         logger,
		metricsHandler: metricsHandler,
		cm:             cm,
	}
}

// StartServing Starts the APIServer
// @author AB
// @param
// @return
func (api *APIServer) StartServing() {
	r := mux.NewRouter()
	r.Handle("/metrics", api.metricsHandler)

	// Validator for Responses
	r.HandleFunc("/ValidateResponse", api.handleValidateResponseMessage).Name("ValidateResponse").Methods("POST")

	//// TODO handlers for specific endpoints were removed as /ValidateResponse will validate all requests
	// for _, element := range configuration.GetEndpointSettings() {
	// 	r.HandleFunc(element.Endpoint, handleMessages).Name(element.Endpoint).Methods("POST")
	// 	log.Log("handling endpoint: "+element.Endpoint, "API", "Main")
	// }

	port := crosscutting.GetEnvironmentValue(api.logger, "API_PORT", ":8080")

	api.logger.Log("Starting the server on port "+port, api.pack, "Main")
	api.logger.Fatal(http.ListenAndServe(port, r), "", api.pack, "Main")
}

// Func: updateReponseError Handles requests to the specified urls in the settings
// @author AB
// @param
// w: Writer to create the response
// genericError: Error to be returned
// responseCode: HTTP Status code
// @return
func (api *APIServer) updateReponseError(w http.ResponseWriter, genericError GenericError, responseCode int) {
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

// handleValidateResponseMessage Handles requests to the specified urls in the settings
// @author AB
// @params
// w: Writer to create the response
// r: Request received
// @return
func (api *APIServer) handleValidateResponseMessage(w http.ResponseWriter, r *http.Request) {
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
		api.updateReponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read the Server Organization ID from the header
	endpointName := r.Header.Get("endpointName")

	// Validate the endpoint configuration exists
	// endpointSettings := settings.GetEndpointSetting(endpointName)
	endpointSettings := api.cm.GetEndpointSettingFromAPI(endpointName, api.logger)

	if endpointSettings == nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "endpointName: Not found or bad format."
		api.updateReponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read header and create a json object
	headerMsg, err := api.buildHeaderMsg(&r.Header)
	if err != nil {
		api.logger.Error(err, "Error: handleValidateResponseMessage", "API", "handleValidateResponseMessage")
		genericError.Message = "Error processing request."
		api.updateReponseError(w, *genericError, http.StatusInternalServerError)
		return
	}

	// Read the body of the message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		genericError.Message = "Failed to read request body."
		api.updateReponseError(w, *genericError, http.StatusInternalServerError)
		return
	}

	// Read the Server Organization ID from the header
	msg.HeaderMessage = headerMsg
	msg.Message = string(body)
	msg.Endpoint = endpointName
	msg.HTTPMethod = r.Method
	msg.ServerID = serverOrgId
	xFapiID := r.Header.Get("x-fapi-interaction-id")
	msg.XFapiInteractionID = xFapiID

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)
	monitoring.RecordResponseDuration(startTime)
	fmt.Fprintf(w, "Message enqueued for processing!")
}

// handleMessages Handles requests to the specified urls in the settings
// @author AB
// @params
// w: Writer to create the response
// r: Request received
// @return
func (api *APIServer) handleMessages(w http.ResponseWriter, r *http.Request) {
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
	headerMsg, err := api.buildHeaderMsg(&r.Header)
	if err != nil {
		api.logger.Error(err, "Error: handleMessages", "API", "handleMessages")
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

// buildHeaderMsg Creates a JSON message based on the headers
// @author AB
// @params
// header: List of headers
// @return
// string: JSON Message created
// error: in case of parsing error
func (api *APIServer) buildHeaderMsg(header *http.Header) (string, error) {
	// Create a map to store header key-value pairs.
	objectMap := make(map[string]interface{})
	headermap := api.getHeaderMap(*header)
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

// headersToJSON Maps all the headers found in the request to a JSON object
// @author AB
// @params
// headers: List of headers to map
// @return
// string: JSON object created
// error: Returs error in case a problem is found during the mapping
func (api *APIServer) getHeaderMap(headers http.Header) map[string]interface{} {
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
