package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients         = make(map[string]*websocket.Conn)
	clientsMux      sync.Mutex
	pendingRequests = make(map[string]chan []byte)
)

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clientID := uuid.NewString()

	addClient(clientID, conn)
	defer removeClient(clientID)

	for {
		// Read the message from the WebSocket
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}

		// Print the received message
		fmt.Printf("Received message: %s\n", p)
		HandleIncomingMessage(p)

	}

}

func HandleIncomingMessage(p []byte) {
	log.Println("Received message from WebSocket client:", string(p))

	// Parse the incoming message
	var messageData map[string]interface{}
	err := json.Unmarshal(p, &messageData)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return
	}

	// Check for the presence of "response_id"
	responseID, ok := messageData["ResponseID"].(string)
	if !ok {
		log.Println("Message does not contain a valid 'response_id':", string(p))
		return
	}

	log.Printf("Processing message with ResponseID: %s\n", responseID)

	clientsMux.Lock()
	defer clientsMux.Unlock()

	// Check if the ResponseID exists in pendingRequests
	ch, exists := pendingRequests[responseID]
	if !exists {
		log.Printf("No pending request found for ResponseID: %s\n", responseID)
		return
	}

	log.Printf("Sending response to channel for ResponseID: %s\n", responseID)

	// Send the response through the channel
	ch <- p

	// Clean up the pending request
	delete(pendingRequests, responseID)
	log.Printf("Removed ResponseID %s from pendingRequests\n", responseID)
}

func addClient(clientID string, conn *websocket.Conn) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	clients[clientID] = conn
	fmt.Printf("Client connected: %s\n", clientID)
}

func removeClient(clientID string) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	delete(clients, clientID)
	fmt.Printf("Client disconnected: %s\n", clientID)
}

func BroadcastMessage(message string) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	for clientID, conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending message to %s: %v", clientID, err)
			conn.Close()
			delete(clients, clientID)
		}
	}
}

func BroadcastMessageToFirstNode(response map[string]interface{}) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	if len(clients) == 0 {
		log.Println("No clients connected!")
		return
	}

	var firstClientId string
	for clientID := range clients {
		firstClientId = clientID
		break
	}

	jsonMessage, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	conn := clients[firstClientId]
	err = conn.WriteMessage(websocket.TextMessage, jsonMessage)
	if err != nil {
		log.Printf("Error sending message to %s: %v", firstClientId, err)
		conn.Close()
		removeClient(firstClientId)
	}
}

func HandleWebSocketRequest(response map[string]interface{}) ([]byte, error) {
	log.Println("HandleWebSocketRequest called")

	responseID := uuid.New().String()
	log.Printf("Generated ResponseID: %s\n", responseID)

	responseChannel := make(chan []byte)

	clientsMux.Lock()
	log.Printf("Adding ResponseID %s to pendingRequests\n", responseID)
	pendingRequests[responseID] = responseChannel
	clientsMux.Unlock()

	response["ResponseID"] = responseID

	log.Println("Broadcasting message to the first WebSocket client")
	BroadcastMessageToFirstNode(response)

	log.Println("Waiting for response from WebSocket client...")
	select {
	case responseData := <-responseChannel:
		log.Printf("Received response for ResponseID %s: %s\n", responseID, string(responseData))
		return responseData, nil
	case <-time.After(30 * time.Second): // Timeout after 30 seconds
		log.Printf("Timeout while waiting for WebSocket response for ResponseID %s\n", responseID)
		clientsMux.Lock()
		delete(pendingRequests, responseID)
		clientsMux.Unlock()
		return nil, fmt.Errorf("timeout waiting for WebSocket response")
	}
}
