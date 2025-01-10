package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"trafficker/websocket"
)

func DynamicRouter(w http.ResponseWriter, r *http.Request) {
	requestMethod := r.Method

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for key, value := range r.Header {
		headers[key] = value[0]
	}

	response := map[string]interface{}{
		"RequestMethod":  requestMethod,
		"FullURL":        r.URL.Path,
		"RequestBody":    string(body),
		"RequestHeaders": headers,
	}

	responseData, err := websocket.HandleWebSocketRequest(response)
	if err != nil {
		log.Println("Error in HandleWebSocketRequest:", err)
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}

	// Parse the response data to verify
	var responsePayload map[string]interface{}
	if err := json.Unmarshal(responseData, &responsePayload); err != nil {
		log.Println("Error unmarshaling WebSocket response:", err)
		http.Error(w, "Failed to process WebSocket response", http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(responsePayload)
	if err != nil {
		log.Println("Error creating JSON response:", err)
		http.Error(w, "Failed to create JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func checkHealth() {
	go func() {
		for {
			log.Println("Health check initiated: Sending status check broadcast...")
			websocket.BroadcastMessage("Status check!")
			time.Sleep(1 * time.Minute)
		}
	}()
}

func main() {
	fmt.Println("Hello, World!")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/api/", DynamicRouter)
	http.HandleFunc("/ws", websocket.HandleConnection)

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	checkHealth()

	for {
		time.Sleep(time.Second)
	}
}
