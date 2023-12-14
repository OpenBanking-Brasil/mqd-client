package worker

import (
	"encoding/json"
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/validation"
	"github.com/OpenBanking-Brasil/MQD_Client/validation/settings"
)

var (
	mutex          = sync.Mutex{} // Mutex for multi processing locks
	singletonMutex = sync.Mutex{}
	singleton      MessageProcessorWorker
)

type MessageProcessorWorker struct {
	pack            string                  // Package name
	logger          log.Logger              // Logger to be used by the package
	receivedValues  map[string]int          // Stores the values for the received messages
	validatedValues map[string]int          // Stores the values for the validated messages
	resultProcessor *result.ResultProcessor // Result processor to be used by the package
	cm              *settings.ConfigurationManager
}

// GetMessageProcessorWorker returns a new message processor
// @author AB
// @params
// logger: Logger to be used by the package
// resultProcessor: Result processor to be used by the package
// @return
// MessageProcessorWorker: New message processor
func GetMessageProcessorWorker(logger log.Logger, resultProcessor *result.ResultProcessor, cm *settings.ConfigurationManager) *MessageProcessorWorker {
	if singleton.pack == "" {
		singletonMutex.Lock()
		defer singletonMutex.Unlock()
		singleton = MessageProcessorWorker{
			pack:            "worker",
			logger:          logger,
			receivedValues:  make(map[string]int),
			validatedValues: make(map[string]int),
			resultProcessor: resultProcessor,
			cm:              cm,
		}
	}

	return &singleton
}

// processMessage Validates and creates a result of a specific message
// @author AB
// @params
// msg: Message to be processed
// @return
func (mp *MessageProcessorWorker) processMessage(msg *queue.Message) {
	mutex.Lock()
	mp.receivedValues[msg.Endpoint]++
	mutex.Unlock()

	// endpointSettings := settings.GetEndpointSetting(msg.Endpoint)
	endpointSettings := mp.cm.GetEndpointSettingFromAPI(msg.Endpoint, mp.logger)

	if endpointSettings == nil {
		mp.logger.Warning("Ignoring message with endpoint: "+msg.Endpoint, "Worker", "processMessage")
	} else {
		vr, err := mp.validateMessage(msg, endpointSettings)
		if err != nil {
			//// TODO handle error
			mp.logger.Error(err, "Error during Validation", "Worker", "processMessage")
		} else {
			// Create a message result entry
			messageResult := result.MessageResult{
				Endpoint:           msg.Endpoint,
				HTTPMethod:         msg.HTTPMethod,
				Result:             vr.Valid,
				Errors:             vr.Errors,
				ServerID:           msg.ServerID,
				XFapiInteractionID: msg.XFapiInteractionID,
			}

			monitoring.IncreaseValidationResult(messageResult.ServerID, messageResult.Endpoint, messageResult.Result)
			mp.resultProcessor.AppendResult(&messageResult)
		}

		mutex.Lock()
		mp.validatedValues[msg.Endpoint]++
		mutex.Unlock()
	}
}

// validateContentWithSchema Validates the content against a specific schema
// @author AB
// @params
// content: Content to be validated
// Schema: String of the JSON schema
// validationResult: Result to be filled with details from the validation
// @return
// Error in case ther is a problem reading or validating the schema
func (mp *MessageProcessorWorker) validateContentWithSchema(content string, schema string, validationResult *validation.ValidationResult) error {
	mp.logger.Info("Validating content with schema", mp.pack, "validateContentWithSchema")
	// Create a dynamic structure from the Message content
	var dynamicStruct validation.DynamicStruct
	err := json.Unmarshal([]byte(content), &dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		return err
	}

	val := validation.GetSchemaValidator(mp.logger, schema)
	valRes, err := val.Validate(dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		mp.logger.Error(err, "Validation error", "Worker", "validateContentWithSchema")
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

// ValidateMessage gets the payload on the message and validates its fields
// @author AB
// @params
// msg: Message to be validated
// settings: Endpoint configuration settings
// @return
// ValidationResult: Result of the validation for the specified message
// error: error in case there is a problem during the validation
// func (mp *MessageProcessorWorker) validateMessage(msg *queue.Message, settings *settings.EndPointSetting) (*validation.ValidationResult, error) {
func (mp *MessageProcessorWorker) validateMessage(msg *queue.Message, settings *settings.APIEndpointSetting) (*validation.ValidationResult, error) {
	mp.logger.Info("Validating message", mp.pack, "validateMessage")
	validationResult := validation.ValidationResult{Valid: true, Errors: make(map[string][]string)}

	err := mp.validateContentWithSchema(msg.HeaderMessage, settings.JSONHeaderSchema, &validationResult)
	if err != nil {
		validationResult.Valid = false
		return &validationResult, err
	}

	err = mp.validateContentWithSchema(msg.Message, settings.JSONBodySchema, &validationResult)
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

// worker is for starting the processing of the queued messages
// @author AB
func (mp *MessageProcessorWorker) worker() {
	for msg := range queue.MessageQueue {
		mp.processMessage(msg)
	}
}

// StartWorker is for starting the worker process
// @author AB
func (mp *MessageProcessorWorker) StartWorker() {
	go mp.worker() // Start the worker Goroutine to process messages

	mp.logger.Log("Worker started.", "Worker", "StartWorker")
}
