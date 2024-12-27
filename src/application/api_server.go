package application

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	xFAPIInteractionID = "x-fapi-interaction-id"
	srvOrgID           = "serverOrgId"
	transmitterID      = "transmitterID"
)

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

	port := as.cm.settings.ConfigurationSettings.APIPort
	// Remove ":" if found
	port = strings.Replace(port, ":", "", -1)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	as.logger.Log("Starting the server on port "+port, as.pack, "StartServing")
	if as.cm.IsHTTPS() {
		as.logger.Fatal(server.ListenAndServeTLS(as.cm.GetCertFilePath(), as.cm.GetKeyFilePath()), "", as.pack, "StartServing")
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

func (as *APIServer) loadMessageHeaderValues(r *http.Request, message *Message) *GenericError {
	genericError := &GenericError{}
	// Read the Server Organization ID from the header
	serverOrgID := r.Header.Get(srvOrgID)
	_, err := uuid.Parse(serverOrgID)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = srvOrgID + ": Not found or bad format."
		return genericError
	}

	xFapiID := r.Header.Get(xFAPIInteractionID)
	_, err = uuid.Parse(xFapiID)
	if err != nil {
		monitoring.IncreaseBadRequestsReceived()
		genericError.Message = xFAPIInteractionID + ": Not found or bad format."
		return genericError
	}

	txServerID := r.Header.Get(transmitterID)
	if txServerID != "" {
		_, err = uuid.Parse(txServerID)
		if err != nil {
			monitoring.IncreaseBadRequestsReceived()
			genericError.Message = transmitterID + ": bad format."
			return genericError
		}
	}

	// Read the Server Organization ID from the header
	endpointName := r.Header.Get("endpointName")

	// Read the api version from the header
	versionHeader := r.Header.Get("version")

	// Read the api version from the header
	consentID := r.Header.Get("consentID")

	message.APIVersion = versionHeader
	message.Endpoint = endpointName
	message.ServerID = serverOrgID
	message.XFapiInteractionID = xFapiID
	message.TransmitterID = txServerID
	message.ConsentID = consentID
	return nil
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

	loadError := as.loadMessageHeaderValues(r, &msg)
	if loadError != nil {
		as.updateResponseError(w, *loadError, http.StatusBadRequest)
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

	// Validate the endpoint configuration exists
	validationSettings := as.cm.GetEndpointSettingFromAPI(msg.Endpoint, as.logger)

	if validationSettings == nil {
		monitoring.IncreaseBadEndpointsReceived(msg.Endpoint, "N.A.", "Endpoint not supported")
		genericError.Message = "endpointName: Not found or bad format."
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	} else if msg.APIVersion != "" && msg.APIVersion != validationSettings.APIVersion {
		monitoring.IncreaseBadEndpointsReceived(msg.Endpoint, msg.APIVersion, "Version not supported")
		genericError.Message = "version: not supported for as endpoint: " + msg.Endpoint
		as.updateResponseError(w, *genericError, http.StatusBadRequest)
		return
	}

	if as.mustValidate(validationSettings.EndpointSettings) {
		msg.Message = string(body)
		msg.HTTPMethod = r.Method

		// Enqueue the message for processing using worker's enqueueMessage
		as.qm.EnqueueMessage(&msg)
	}

	monitoring.RecordResponseDuration(startTime)
	_, err = fmt.Fprintf(w, "Message enqueued for processing!")
	if err != nil {
		as.logger.Error(err, "Error writing response:", as.pack, "handleValidateResponseMessage")
	}
}
