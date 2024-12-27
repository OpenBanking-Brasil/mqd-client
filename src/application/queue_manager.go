package application

import (
	"encoding/json"

	"github.com/OpenBanking-Brasil/MQD_Client/validation"
)

// Message contains the information of the Payload to be validated
type Message struct {
	Message string `json:"message"` // Body Payload sent to the API
	//HeaderMessage      string `json:"header_message"` // Header Payload sent to the API
	Endpoint           string `json:"endpoint"`    // Name of the endpoint requested
	APIVersion         string `json:"api_version"` // Version of the API to validate
	HTTPMethod         string `json:"http_method"` // HTTP Method used
	ServerID           string `json:"server_id"`   // Identifier of the Client requesting the information
	XFapiInteractionID string
	ConsentID          string
	TransmitterID      string // Organisation ID of the transmitter
}

// GetMappedObject Returns the json message object mapped as a dynamic structure
//
// Parameters:
//
// Returns:
//   - DynamicStruct: mapped object
//   - Error: Error if any during the unmarshal process
func (msg *Message) GetMappedObject() (validation.DynamicStruct, error) {
	// Create a dynamic structure from the Message content
	var dynamicStruct validation.DynamicStruct
	err := json.Unmarshal([]byte(msg.Message), &dynamicStruct)
	if err != nil {
		return nil, err
	}

	return dynamicStruct, nil
}

// Buffered channel for message queue
var messageQueue = make(chan *Message, 1000)

// QueueManager is in charge of managing the queue for messages to process
type QueueManager struct {
}

// GetQueueManager returns a new queue manager
//
// Parameters:
//
// Returns:
//   - *QueueManager: New queue manager
func GetQueueManager() *QueueManager {
	return &QueueManager{}
}

// EnqueueMessage is for queueing the message
//
// Parameters:
//   - msg: Message to be queued
//
// Returns:
func (qm *QueueManager) EnqueueMessage(msg *Message) {
	messageQueue <- msg
}

// GetQueue returns the list of messages in the queue
//
// Parameters:
//
// Returns:
//   - chan *Message: List of messages in the queue
func (qm *QueueManager) GetQueue() chan *Message {
	return messageQueue
}
