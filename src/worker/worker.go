package worker

import (
	"encoding/json"
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/validator"
)

var (
	receivedValues  = make(map[string]int) // Stores the values for the received messages
	validatedValues = make(map[string]int) // Stores the values for the validated messages
	mutex           = sync.Mutex{}         // Mutex for multi processing locks
)

// Func: processMessage Validates and creates a result of a specific message
// @author AB
// @params
// msg: MEssage to be processed
// @return
func processMessage(msg *queue.Message) {
	mutex.Lock()
	receivedValues[msg.Endpoint]++
	mutex.Unlock()

	endpointSettings := configuration.GetEndpointSetting(msg.Endpoint)

	if endpointSettings.Endpoint == "" {
		log.Warning("Ignoring message with endpoint: "+msg.Endpoint, "Worker", "processMessage")
	} else {
		vr, err := validateMessage(msg, endpointSettings)
		if err != nil {
			//// TODO handle error
			log.Error(err, "Error during Validation", "Worker", "processMessage")
		} else {
			// Create a message result entry
			messageResult := result.MessageResult{
				Endpoint:   msg.Endpoint,
				HTTPMethod: msg.HTTPMethod,
				Result:     vr.Valid,
				Errors:     vr.Errors,
				ServerID:   msg.ServerID,
			}

			monitoring.IncreaseValidationResult(messageResult.ServerID, messageResult.Endpoint, messageResult.Result)
			result.AppendResult(&messageResult)
		}

		mutex.Lock()
		validatedValues[msg.Endpoint]++
		mutex.Unlock()
	}
}

// Func: validateContentWithSchema Validates the content against a specific schema
// @author AB
// @params
// content: Content to be validated
// Schema: String of the JSON schema
// validationResult: Result to be filled with details from the validation
// @return
// Error in case ther is a problem reading or validating the schema
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
		log.Error(err, "Validation error", "Worker", "validateContentWithSchema")
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

// Func: ValidateMessage gets the payload on the message and validates its fields
// @author AB
// @params
// msg: Message to be validated
// settings: Endpoint configuration settings
// @return
// ValidationResult: Result of the validation for the specified message
// error: error in case there is a problem during the validation
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
	//// rules, err := validator.LoadValidationRules("ParameterData\\validation_rules.json")
	//// if err != nil {
	//// 	validationResult.Valid = false
	//// 	return validationResult, err
	//// }

	return &validationResult, nil
}

// Func: worker is for starting the processing of the queued messages
// @author AB
func worker() {
	for msg := range queue.MessageQueue {
		processMessage(msg)
	}
}

// Func: StartWorker is for starting the worker process
// @author AB
func StartWorker() {
	go worker() // Start the worker Goroutine to process messages

	log.Log("Worker started.", "Worker", "StartWorker")
}
