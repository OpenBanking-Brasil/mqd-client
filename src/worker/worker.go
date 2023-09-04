package worker

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/validator"
)

var (
	receivedValues  = make(map[string]int)
	validatedValues = make(map[string]int)
	mutex           = sync.Mutex{}
)

/**
 * Func: processMessage Validates and creates a result of a specific message
 *
 * @author AB
 *
 * @params
 * msg: MEssage to be processed
 * @return
 */
func processMessage(msg *queue.Message) {
	// Update received value for the endpoint
	mutex.Lock()
	receivedValues[msg.Endpoint]++
	mutex.Unlock()

	endpointSettings := getEndpointSettings(msg.Endpoint)

	if endpointSettings.Endpoint == "" {
		fmt.Printf("Ignoring message with endpoint: %s\n", msg.Endpoint)
	} else {
		vr := ValidateMessage(msg, endpointSettings)
		// fmt.Printf("Valid: %s, ErrorType: %s\n", vr.Valid, vr.ErrType)
		// Create a message result entry
		messageResult := result.MessageResult{
			Endpoint:   msg.Endpoint,
			HTTPMethod: msg.HTTPMethod,
			Result:     vr.Valid,
			ClientID:   msg.ClientID,
			ServerID:   msg.ServerID,
		}

		result.AppendResult(&messageResult)

		// Here, you can define the validation logic for the received message.
		// For this example, let's assume it's valid and just update the validated value.
		mutex.Lock()
		validatedValues[msg.Endpoint]++
		mutex.Unlock()
	}
}

func getEndpointSettings(endpointName string) *configuration.EndPointSettings {
	for _, element := range configuration.GetEndpointSettings() {
		if element.Endpoint == endpointName {
			return &element
		}
	}

	return &configuration.EndPointSettings{}
}

/**
 * Func: ValidateMessage gets the payload on the message and validates its fields
 *
 * @author AB
 *
 * @params
 * msg: Message to be validated
 * @return
 * ValidationResult: Result of the validation for the specified message
 */
func ValidateMessage(msg *queue.Message, settings *configuration.EndPointSettings) validator.ValidationResult {
	println("Validating message")
	validationResult := validator.ValidationResult{Valid: true}

	// Load validation rules from the CSV file
	rules, err := validator.LoadValidationRules("ParameterData\\validation_rules.json")
	if err != nil {
		validationResult.Valid = false
		validationResult.ErrType = "Validation error: " + err.Error()
		return validationResult
	}

	// Create a dynamic structure from the Message content
	var headerDynamicStruct validator.DynamicStruct
	err = json.Unmarshal([]byte(msg.HeaderMessage), &headerDynamicStruct)
	if err != nil {
		// http.Error(w, "Invalid dynamic structure JSON", http.StatusBadRequest)
		validationResult.Valid = false
		validationResult.ErrType = err.Error()
		return validationResult
	}

	val := validator.NewValidator(rules)

	err = val.ValidateWithSchema(headerDynamicStruct, settings)
	if err != nil {
		validationResult.Valid = false
		validationResult.ErrType = "Validation error: " + err.Error()
		println("Validation error: " + err.Error())
	}

	return validationResult
}

/**
 * Func: worker is for starting the processing of the queued messages
 *
 * @author AB
 */
func worker() {
	for msg := range queue.MessageQueue {
		processMessage(msg)
	}
}

/**
 * Func: StartWorker is for starting the worker process
 *
 * @author AB
 */

func StartWorker() {
	go worker() // Start the worker Goroutine to process messages

	fmt.Println("Worker started.")
}
