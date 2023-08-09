package worker

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/validator"
)

var (
	endpointsToProcess = map[string]struct{}{
		"/endpoint1":   {},
		"/endpoint2":   {},
		"Endpoint URL": {},
		// Add more endpoints here as needed
	} // List of Endpoints that MUST be validated

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

	if _, ok := endpointsToProcess[msg.Endpoint]; !ok {
		fmt.Printf("Ignoring message with endpoint: %s\n", msg.Endpoint)
	} else {
		vr := ValidateMessage(msg)
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
func ValidateMessage(msg *queue.Message) validator.ValidationResult {
	validationResult := validator.ValidationResult{Valid: true}

	// Create a dynamic structure from the Message content
	var dynamicStruct validator.DynamicStruct
	err := json.Unmarshal([]byte(msg.Message), &dynamicStruct)
	if err != nil {
		// http.Error(w, "Invalid dynamic structure JSON", http.StatusBadRequest)
		validationResult.Valid = false
		validationResult.ErrType = err.Error()
		return validationResult
	}

	// Load validation rules from the CSV file
	rules, err := validator.LoadValidationRules("..\\validation_rules.csv")
	if err != nil {
		validationResult.ErrType = err.Error()
		return validationResult
	}

	validator := validator.NewValidator(rules)

	err = validator.Validate(dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		validationResult.ErrType = "Validation error: " + err.Error()
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
