package application

import (
	"encoding/json"
	"sync"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/domain/models"
	"github.com/OpenBanking-Brasil/MQD_Client/validation"
)

var (
	messageProcessorWorkerMutex = sync.Mutex{}          // Mutex for multiprocessing locks
	singletonMutex              = sync.Mutex{}          // Mutex for the singleton variable
	messageProcessorSingleton   *MessageProcessorWorker // Message process singleton
)

// MessageProcessorWorker is in charge of processing the message requests
type MessageProcessorWorker struct {
	crosscutting.OFBStruct
	receivedValues  map[string]int        // Stores the values for the received messages
	validatedValues map[string]int        // Stores the values for the validated messages
	resultProcessor *ResultProcessor      // Result processor to be used by the package
	cm              *ConfigurationManager // Configuration manager
	qm              *QueueManager         // Queue manager to queue the messages
	lrm             *LocalResultManager
}

// GetMessageProcessorWorker returns a new message processor
//
// Parameters:
//   - logger: Logger to be used by the package
//   - resultProcessor: Result processor to be used by the package
//   - qm: Queue manager
//   - cm: Configuration manager
//
// Returns:
//   - MessageProcessorWorker: New message processor
func GetMessageProcessorWorker(logger log.Logger, resultProcessor *ResultProcessor, qm *QueueManager, cm *ConfigurationManager, lrm *LocalResultManager) *MessageProcessorWorker {
	if messageProcessorSingleton == nil {
		singletonMutex.Lock()
		defer singletonMutex.Unlock()
		messageProcessorSingleton = &MessageProcessorWorker{
			OFBStruct: crosscutting.OFBStruct{
				Pack:   "worker",
				Logger: logger,
			},

			receivedValues:  make(map[string]int),
			validatedValues: make(map[string]int),
			resultProcessor: resultProcessor,
			qm:              qm,
			cm:              cm,
			lrm:             lrm,
		}
	}

	return messageProcessorSingleton
}

// processMessage Validates and creates a result of a specific message
//
// Parameters:
//   - msg: Message to be processed
//
// Returns:
func (mpw *MessageProcessorWorker) processMessage(msg *Message) {
	messageProcessorWorkerMutex.Lock()
	mpw.receivedValues[msg.Endpoint]++
	messageProcessorWorkerMutex.Unlock()

	validationSettings := mpw.cm.GetEndpointSettingFromAPI(msg.Endpoint, mpw.Logger)

	if validationSettings == nil {
		mpw.Logger.Warning("Ignoring message with endpoint: "+msg.Endpoint, mpw.Pack, "processMessage")
	} else {
		messageResult := MessageResult{
			Endpoint:           msg.Endpoint,
			HTTPMethod:         msg.HTTPMethod,
			ServerID:           msg.ServerID,
			XFapiInteractionID: msg.XFapiInteractionID,
			TransmitterID:      msg.TransmitterID,
		}
		if msg.ConsentID != "" {
			messageResult.XFapiInteractionID = "[" + msg.ConsentID + "] - [" + msg.XFapiInteractionID + "]"
		}

		vr, err := mpw.validateMessage(msg, validationSettings.EndpointSettings)
		if err != nil {
			mpw.Logger.Error(err, "Error during Validation for endpoint: "+msg.Endpoint, mpw.Pack, "processMessage")
			messageResult.Result = false
			messageResult.Errors = map[string][]string{
				"(error)": {err.Error()},
			}
		} else {
			// Create a message result entry
			messageResult.Result = vr.Valid
			messageResult.Errors = vr.Errors
		}

		monitoring.IncreaseValidationResult(messageResult.ServerID, messageResult.Endpoint, messageResult.Result)
		mpw.resultProcessor.AppendResult(&messageResult)
		mpw.lrm.AppendResult(*msg, messageResult, *validationSettings)
		messageProcessorWorkerMutex.Lock()
		mpw.validatedValues[msg.Endpoint]++
		messageProcessorWorkerMutex.Unlock()
	}
}

// validateContentWithSchema Validates the content against a specific schema
//
// Parameters:
//   - content: Content to be validated
//   - schema: JSON schema to validate with
//   - validationResult: Result to be filled with details from the validation
//
// Returns:
//   - error: Error in case there is a problem reading or validating the schema
func (mpw *MessageProcessorWorker) validateContentWithSchema(content string, schema string, validationResult *validation.Result) error {
	mpw.Logger.Info("Validating content with schema", mpw.Pack, "validateContentWithSchema")

	// Create a dynamic structure from the Message content
	var dynamicStruct validation.DynamicStruct
	err := json.Unmarshal([]byte(content), &dynamicStruct)
	if err != nil {
		mpw.Logger.Error(err, "Error unmarshalling content", mpw.Pack, "validateContentWithSchema")
		mpw.Logger.Debug("Content message: "+content, mpw.Pack, "validateContentWithSchema")
		validationResult.Valid = false
		return err
	}

	val := validation.GetSchemaValidator(mpw.Logger, schema)
	valRes, err := val.Validate(dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		mpw.Logger.Error(err, "Validation error", mpw.Pack, "validateContentWithSchema")
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

// ValidateMessage gets the Payload on the message and validates its fields
//
// Parameters:
//   - msg: Message to be validated
//   - settings: Endpoint configuration settings
//
// Returns:
//   - ValidationResult: Result of the validation for the specified message
//   - error: error in case there is a problem during the validation
func (mpw *MessageProcessorWorker) validateMessage(msg *Message, settings *models.APIEndpointSetting) (*validation.Result, error) {
	mpw.Logger.Info("Validating message for endpoint: "+msg.Endpoint, mpw.Pack, "validateMessage")
	validationResult := validation.Result{Valid: true, Errors: make(map[string][]string)}

	err := mpw.validateContentWithSchema(msg.Message, settings.JSONBodySchema, &validationResult)
	if err != nil {
		mpw.Logger.Error(err, "Error during body validation", mpw.Pack, "validateMessage")
		validationResult.Valid = false
		return &validationResult, err
	}

	return &validationResult, nil
}

// worker is for starting the processing of the queued messages
//
// Parameters:
//
// Returns:
func (mpw *MessageProcessorWorker) worker() {
	for msg := range mpw.qm.GetQueue() {
		mpw.processMessage(msg)
	}
}

// StartWorker is for starting the worker process
//
// Parameters:
//
// Returns:
func (mpw *MessageProcessorWorker) StartWorker() {
	go mpw.worker() // Start the worker Goroutine to process messages

	mpw.Logger.Log("Worker started.", mpw.Pack, "StartWorker")
}
