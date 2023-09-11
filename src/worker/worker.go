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

	endpointSettings := configuration.GetEndpointSetting(msg.Endpoint)

	if endpointSettings.Endpoint == "" {
		fmt.Printf("Ignoring message with endpoint: %s\n", msg.Endpoint)
	} else {
		vr, err := validateMessage(msg, endpointSettings)
		if err != nil {
			//// TODO handle error
			println("Error validating!! " + err.Error())
		} else {
			// Create a message result entry
			messageResult := result.MessageResult{
				Endpoint:   msg.Endpoint,
				HTTPMethod: msg.HTTPMethod,
				Result:     vr.Valid,
				Errors:     vr.Errors,
				ServerID:   msg.ServerID,
			}

			result.AppendResult(&messageResult)
		}

		// Here, you can define the validation logic for the received message.
		// For this example, let's assume it's valid and just update the validated value.
		mutex.Lock()
		validatedValues[msg.Endpoint]++
		mutex.Unlock()
	}
}

func validateContentWithSchema(content string, schema string, validationResult *validator.ValidationResult) error {
	// Create a dynamic structure from the Message content
	var dynamicStruct validator.DynamicStruct
	err := json.Unmarshal([]byte(content), &dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		return err
	}

	val := validator.NewValidator()

	valRes, err := val.ValidateWithSchema(dynamicStruct, schema)
	if err != nil {
		validationResult.Valid = false
		println("Validation error: " + err.Error())
		return err
	}

	if !valRes.Valid {
		for key, value := range valRes.Errors {
			validationResult.Errors[key] = value
		}

		validationResult.Valid = valRes.Valid
	}

	return nil
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
func validateMessage(msg *queue.Message, settings *configuration.EndPointSettings) (*validator.ValidationResult, error) {
	validationResult := validator.ValidationResult{Valid: true, Errors: make(map[string][]string)}

	err := validateContentWithSchema(msg.HeaderMessage, settings.JSONHeaderSchema, &validationResult)
	if err != nil {
		validationResult.Valid = false
		return &validationResult, err
	}

	err = validateContentWithSchema(msg.Message, settings.JSONBodySchema, &validationResult)
	if err != nil {
		validationResult.Valid = false
		return &validationResult, err
	}

	// // Load validation rules from the CSV file
	// rules, err := validator.LoadValidationRules("ParameterData\\validation_rules.json")
	// if err != nil {
	// 	validationResult.Valid = false
	// 	return validationResult, err
	// }

	return &validationResult, nil
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
