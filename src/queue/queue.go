package queue

// Message contains the information of the payload to be validated
type Message struct {
	Message       string `json:"message"`        // Payload sent to the API
	HeaderMessage string `json:"header_message"` // Payload sent to the API
	Endpoint      string `json:"endpoint"`       // Name of the endpoint requested
	HTTPMethod    string `json:"http_method"`    // HTTP Method used
	ClientID      string `json:"client_id"`      // Identifier of the Client requesting the information
	ServerID      string `json:"server_id"`      // Identifies the server requesting the information
}

// Buffered channel for message queue
var MessageQueue = make(chan *Message, 1000)

/**
 * Func: EnqueueMessage is for queueing the message
 *
 * @author AB
 *
 * @params
 * msg: Message to be queued
 * @return
 */
func EnqueueMessage(msg *Message) {
	MessageQueue <- msg
}
