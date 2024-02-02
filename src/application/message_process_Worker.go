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
	messageProcessorWorkerMutex = sync.Mutex{} // Mutex for multi processing locks
	singletonMutex              = sync.Mutex{}
	messageProcessorSingleton   *MessageProcessorWorker
)

type MessageProcessorWorker struct {
	crosscutting.OFBStruct
	receivedValues  map[string]int   // Stores the values for the received messages
	validatedValues map[string]int   // Stores the values for the validated messages
	resultProcessor *ResultProcessor // Result processor to be used by the package
	cm              *ConfigurationManager
	qm              *QueueManager
}

// GetMessageProcessorWorker returns a new message processor
// @author AB
// @params
// logger: Logger to be used by the package
// resultProcessor: Result processor to be used by the package
// @return
// MessageProcessorWorker: New message processor
func GetMessageProcessorWorker(logger log.Logger, resultProcessor *ResultProcessor, qm *QueueManager, cm *ConfigurationManager) *MessageProcessorWorker {
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
		}
	}

	return messageProcessorSingleton
}

// processMessage Validates and creates a result of a specific message
// @author AB
// @params
// msg: Message to be processed
// @return
func (this *MessageProcessorWorker) processMessage(msg *Message) {
	messageProcessorWorkerMutex.Lock()
	this.receivedValues[msg.Endpoint]++
	messageProcessorWorkerMutex.Unlock()

	// endpointSettings := settings.GetEndpointSetting(msg.Endpoint)
	endpointSettings, _ := this.cm.GetEndpointSettingFromAPI(msg.Endpoint, this.Logger)

	if endpointSettings == nil {
		this.Logger.Warning("Ignoring message with endpoint: "+msg.Endpoint, this.Pack, "processMessage")
	} else {
		vr, err := this.validateMessage(msg, endpointSettings)
		if err != nil {
			//// TODO handle error
			this.Logger.Error(err, "Error during Validation", this.Pack, "processMessage")
		} else {
			// Create a message result entry
			messageResult := MessageResult{
				Endpoint:           msg.Endpoint,
				HTTPMethod:         msg.HTTPMethod,
				Result:             vr.Valid,
				Errors:             vr.Errors,
				ServerID:           msg.ServerID,
				XFapiInteractionID: msg.XFapiInteractionID,
			}

			monitoring.IncreaseValidationResult(messageResult.ServerID, messageResult.Endpoint, messageResult.Result)
			this.resultProcessor.AppendResult(&messageResult)
		}

		messageProcessorWorkerMutex.Lock()
		this.validatedValues[msg.Endpoint]++
		messageProcessorWorkerMutex.Unlock()
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
func (this *MessageProcessorWorker) validateContentWithSchema(content string, schema string, validationResult *validation.ValidationResult) error {
	this.Logger.Info("Validating content with schema", this.Pack, "validateContentWithSchema")
	// Create a dynamic structure from the Message content
	var dynamicStruct validation.DynamicStruct
	err := json.Unmarshal([]byte(content), &dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		return err
	}

	val := validation.GetSchemaValidator(this.Logger, schema)
	valRes, err := val.Validate(dynamicStruct)
	if err != nil {
		validationResult.Valid = false
		this.Logger.Error(err, "Validation error", this.Pack, "validateContentWithSchema")
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
// func (this *MessageProcessorWorker) validateMessage(msg *queue.Message, settings *settings.EndPointSetting) (*validation.ValidationResult, error) {
func (this *MessageProcessorWorker) validateMessage(msg *Message, settings *models.APIEndpointSetting) (*validation.ValidationResult, error) {
	this.Logger.Info("Validating message", this.Pack, "validateMessage")
	validationResult := validation.ValidationResult{Valid: true, Errors: make(map[string][]string)}

	err := this.validateContentWithSchema(msg.HeaderMessage, settings.JSONHeaderSchema, &validationResult)
	if err != nil {
		validationResult.Valid = false
		return &validationResult, err
	}

	err = this.validateContentWithSchema(msg.Message, settings.JSONBodySchema, &validationResult)
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
func (this *MessageProcessorWorker) worker() {
	for msg := range this.qm.GetQueue() {
		this.processMessage(msg)
	}
}

// StartWorker is for starting the worker process
// @author AB
func (this *MessageProcessorWorker) StartWorker() {
	go this.worker() // Start the worker Goroutine to process messages

	this.Logger.Log("Worker started.", this.Pack, "StartWorker")
}
