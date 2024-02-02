package application

// Message contains the information of the payload to be validated
type Message struct {
	Message            string `json:"message"`        // Body Payload sent to the API
	HeaderMessage      string `json:"header_message"` // Header Payload sent to the API
	Endpoint           string `json:"endpoint"`       // Name of the endpoint requested
	HTTPMethod         string `json:"http_method"`    // HTTP Method used
	ServerID           string `json:"server_id"`      // Identifier of the Client requesting the information
	XFapiInteractionID string
}

// Buffered channel for message queue
var messageQueue = make(chan *Message, 1000)

type QueueManager struct {
}

func GetQueueManager() *QueueManager {
	return &QueueManager{}
}

// EnqueueMessage is for queueing the message
// @author AB
// @params
// msg: Message to be queued
// @return
func (q *QueueManager) EnqueueMessage(msg *Message) {
	messageQueue <- msg
}

func (q *QueueManager) GetQueue() chan *Message {
	return messageQueue
}
