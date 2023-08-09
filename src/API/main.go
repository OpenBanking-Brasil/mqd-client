// api/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/OpenBanking-Brasil/MQD_Client/monitoring"
	"github.com/OpenBanking-Brasil/MQD_Client/queue"
	"github.com/OpenBanking-Brasil/MQD_Client/result"
	"github.com/OpenBanking-Brasil/MQD_Client/worker" // Import the worker package
	// Other imports...
)

func handleMessages(w http.ResponseWriter, r *http.Request) {
	monitoring.Requests.Add(context.Background(), 1)
	var msg queue.Message

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Enqueue the message for processing using worker's enqueueMessage
	queue.EnqueueMessage(&msg)

	fmt.Fprintf(w, "Message enqueued for processing!")
}

func main() {
	monitoring.StartOpenTelemetry()

	// Start the worker Goroutine to process messages
	go worker.StartWorker()
	go result.StartResultsProcessor()

	http.HandleFunc("/sendmessage", handleMessages)

	fmt.Println("Starting the server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
