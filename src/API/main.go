// api/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/worker"
	"github.com/gorilla/mux"
)

func handleMessages(w http.ResponseWriter, r *http.Request) {
	monitoring.Requests.Add(context.Background(), 1)
	var msg queue.Message

	routeName := mux.CurrentRoute(r).GetName()
	// decoder := json.NewDecoder(r.Body)
	// err := decoder.Decode(&msg)
	// if err != nil {
	// 	http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
	// 	return
	// }

	// Convert headers to JSON.
	jsonData, err := headersToJSON(r.Header)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	msg.HeaderMessage = jsonData
	msg.Endpoint = routeName
	msg.HTTPMethod = r.Method

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)

	fmt.Fprintf(w, "Message enqueued for processing!")
}

func main() {
	monitoring.StartOpenTelemetry()

	configuration.Initialize()
	// Start the worker Goroutine to process messages
	go worker.StartWorker()
	go result.StartResultsProcessor()

	r := mux.NewRouter()

	r.HandleFunc("/sendmessage", handleMessages).Name("/sendmessage").Methods("GET")
	for _, element := range configuration.GetEndpointSettings() {
		r.HandleFunc(element.Endpoint, handleMessages).Name(element.Endpoint).Methods("GET")
		println("handling endpoint: " + element.Endpoint)
	}

	//http.Handle("/", r)
	port := crosscutting.GetEnvironmentValue("API_PORT", ":8080")

	fmt.Println("Starting the server on port " + port)
	log.Fatal(http.ListenAndServe(port, r))
}

func headersToJSON(headers http.Header) (string, error) {
	// Create a map to store header key-value pairs.
	headerMap := make(map[string]interface{})

	// Iterate through the header parameters and add them to the map.
	for key, values := range headers {
		// If there's only one value for the header, store it directly.
		if len(values) == 1 {
			headerMap[key] = values[0]
		} else {
			// If there are multiple values, store them as an array.
			headerMap[key] = values
		}
	}

	// Convert the map to a JSON string.
	jsonData, err := json.Marshal(headerMap)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
