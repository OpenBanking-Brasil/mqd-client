package application

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

//const ENV_CLIENT_CRT_FILE = "CLIENT_CRT_FILE" // Certificate file
//const ENV_CLIENT_KEY_FILE = "CLIENT_KEY_FILE" // Private key file

// GenericError contains information message when error needs to be returned
type GenericError struct {
	Message string // Error message
}

// APIServer Contains the APIServer
type APIServer struct {
	pack           string                // Package name
	logger         log.Logger            // Logger to be used
	metricsHandler http.Handler          // Handler for the metric endpoint
	qm             *QueueManager         // Manager for the message queue
	cm             *ConfigurationManager // Manager for application settings
}

// GetAPIServer Creates a new APIServer
//
// Parameters:
//   - logger: Logger to be used
//   - metricsHandler: Metric handler to expose \metrics
//   - qm: Queue manager to queue the requests
//   - cm: ConfigurationManager to handle the configuration
//
// Returns:
//   - *APIServer: APIServer created
func GetAPIServer(logger log.Logger, metricsHandler http.Handler, qm *QueueManager, cm *ConfigurationManager) *APIServer {
	return &APIServer{
		pack:           "API",
		logger:         logger,
		metricsHandler: metricsHandler,
		qm:             qm,
		cm:             cm,
	}
}

// StartServing Starts the APIServer
//
// Parameters:
// Returns:
func (as *APIServer) StartServing() {
	r := mux.NewRouter()
	r.Handle("/metrics", as.metricsHandler)

	// Validator for Responses
	r.HandleFunc("/ValidateResponse", as.handleValidateResponseMessage).Name("ValidateResponse").Methods("POST")

	port := crosscutting.GetEnvironmentValue(as.logger, "API_PORT", ":8080")

	server := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	as.logger.Log("Starting the server on port "+port, as.pack, "StartServing")
	if as.cm.EnableHTTPS {
		as.logger.Fatal(server.ListenAndServeTLS(as.cm.CertFilePath, as.cm.KeyFilePath), "", as.pack, "StartServing")
	} else {
		as.logger.Fatal(server.ListenAndServe(), "", as.pack, "StartServing")
	}
}

// updateResponseError Handles requests to the specified urls in the settings
//
// Parameters:
//   - w: Writer to create the response
//   - genericError: genericError with the error information
//   - responseCode: HTTP response code
//
// Returns:
func (as *APIServer) updateResponseError(w http.ResponseWriter, genericError GenericError, responseCode int) {
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
	_, err = w.Write(jsonData)
	if err != nil {
		as.logger.Error(err, "Error writing JSON response:", as.pack, "updateResponseError")
		return
	}
}

// mustValidate indicates if the endpoint should be validated or not base on the validation rate configured
//
// Parameters:
//   - endpointSettings: Endpoint settings with the configuration information
//
// Returns:
//   - bool: true if the endpoint should be validated
func (as *APIServer) mustValidate(endpointSetting *models.APIEndpointSetting) bool {
	value := as.getRandomNumber()
	switch endpointSetting.Throughput {
	case models.ExtremelyHighTroughput:
		return value < as.cm.ConfigurationSettings.ValidationSettings.ExtremelyHighTroughputValidationRate
	case models.HighTroughput:
		return value < as.cm.ConfigurationSettings.ValidationSettings.HighTroughputValidationRate
	case models.MediumTroughput:
		return value < as.cm.ConfigurationSettings.ValidationSettings.MediumTroughputValidationRate
	case models.LowTroughput:
		return value < as.cm.ConfigurationSettings.ValidationSettings.LowTroughputValidationRate
	case models.VeryLowTroughput:
		return value < as.cm.ConfigurationSettings.ValidationSettings.VeryLowTroughputValidationRate
	}

	return true
}

// getRandomNumber generates a new random number using Cryptographic Randomness
//
// Returns:
//   - int: Random number generated
func (as *APIServer) getRandomNumber() int {
	// Define the upper limit (101 for inclusive range of 0-100)
	maxRandomNumber := big.NewInt(101)

	// Generate a random number between 0 (inclusive) and maxRandomNumber (exclusive)
	num, err := rand.Int(rand.Reader, maxRandomNumber)
	if err != nil {
		as.logger.Error(err, "Error generating random number:", as.pack, "getRandomNumber")
		return 100
	}

	// Convert the big.Int to an int for easier use
	number := int(num.Int64())

	return number
}

// handleValidateResponseMessage Handles requests to the specified urls in the settings
//
// Parameters:
//   - w: Writer to create the response
//   - r: Request received
//
// Returns:
func (as *APIServer) handleValidateResponseMessage(w http.ResponseWriter, r *http.Request) {
	genericError := &GenericError{}
	startTime := time.Now()
	monitoring.IncreaseRequestsReceived()
	var msg Message

	// Read the Server Organization ID from the header
	serverOrgID := r.Header.Get("serverOrgId")
	_, err := uuid.Parse(serverOrgID)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "serverOrgId: Not found or bad format."
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	xFapiID := r.Header.Get("x-fapi-interaction-id")
	_, err = uuid.Parse(xFapiID)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "x-fapi-interaction-id: Not found or bad format."
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read the body of the message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		genericError.Message = "Failed to read request body."
		as.updateResponseError(w, *genericError, http.StatusInternalServerError)
		return
	}

	var js json.RawMessage
	validJSON := json.Unmarshal(body, &js) == nil
	if !validJSON {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = "body: Not a Valid JSON Message."
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	// Read the Server Organization ID from the header
	endpointName := r.Header.Get("endpointName")

	// Validate the endpoint configuration exists
	endpointSettings, version := as.cm.GetEndpointSettingFromAPI(endpointName, as.logger)
	// Read the api version from the header
	versionHeader := r.Header.Get("version")

	if endpointSettings == nil {
		monitoring.IncreaseBadEndpointsReceived(endpointName, "N.A.", "Endpoint not supported")
		genericError.Message = "endpointName: Not found or bad format."
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	} else if versionHeader != "" && versionHeader != version {
		monitoring.IncreaseBadEndpointsReceived(endpointName, versionHeader, "Version not supported")
		genericError.Message = "version: not supported for as endpoint: " + endpointName
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	if as.mustValidate(endpointSettings) {
		// Read header and create a json object
		headerMsg, err := as.buildHeaderMsg(&r.Header)
		if err != nil {
			as.logger.Error(err, "Error: handleValidateResponseMessage", "API", "handleValidateResponseMessage")
			genericError.Message = "Error processing request."
			as.updateResponseError(w, *genericError, http.StatusInternalServerError)
			return
		}

		// Read the Server Organization ID from the header
		msg.HeaderMessage = headerMsg
		msg.Message = string(body)
		msg.Endpoint = endpointName
		msg.HTTPMethod = r.Method
		msg.ServerID = serverOrgID

		msg.XFapiInteractionID = xFapiID

		// Enqueue the message for processing using worker's enqueueMessage
		as.qm.EnqueueMessage(&msg)
	}

	monitoring.RecordResponseDuration(startTime)
	_, err = fmt.Fprintf(w, "Message enqueued for processing!")
	if err != nil {
		as.logger.Error(err, "Error writing response:", as.pack, "handleValidateResponseMessage")
	}
}

// buildHeaderMsg Creates a JSON message based on the headers
//
// Parameters:
//   - header: List of headers
//
// Returns:
//   - string: JSON Message created
//   - error: in case of parsing error
func (as *APIServer) buildHeaderMsg(header *http.Header) (string, error) {
	headerMap := as.getHeaderMap(*header)

	// Convert the map to a JSON string.
	jsonData, err := json.Marshal(headerMap)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// headersToJSON Maps all the headers found in the request to a JSON object
//
// Parameters:
//   - headers: List of headers to map
//
// Returns:
//   - map: map[string]interface{} that contains the list of headers and its values
func (as *APIServer) getHeaderMap(headers http.Header) map[string]interface{} {
	// Create a map to store header key-value pairs.
	headerMap := make(map[string]interface{})

	// Iterate through the header parameters and add them to the map.
	for key, values := range headers {
		key = strings.ToLower(key)
		// Sanitize each header value using html.EscapeString (adjust based on data type)
		sanitizedValues := make([]string, 0, len(values))
		for _, value := range values {
			sanitizedValues = append(sanitizedValues, html.EscapeString(value))
		}
		headerMap[key] = sanitizedValues
	}

	return headerMap
}
