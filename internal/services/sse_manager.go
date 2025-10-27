package services

import (
	"comment-review-platform/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// SSEManager manages Server-Sent Events connections
type SSEManager struct {
	clients   map[int]chan string // userID -> message channel
	mu        sync.RWMutex
	broadcast chan models.BroadcastMessage
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &SSEManager{
		clients:   make(map[int]chan string),
		broadcast: make(chan models.BroadcastMessage, 100),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start the broadcast worker
	go manager.broadcastWorker()

	// Start heartbeat worker
	go manager.heartbeatWorker()

	return manager
}

// AddClient adds a new SSE client connection
func (m *SSEManager) AddClient(userID int) chan string {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a buffered channel for this client
	clientChan := make(chan string, 10)
	m.clients[userID] = clientChan

	log.Printf("SSE client added for user %d, total clients: %d", userID, len(m.clients))
	return clientChan
}

// RemoveClient removes a client connection
func (m *SSEManager) RemoveClient(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clientChan, exists := m.clients[userID]; exists {
		close(clientChan)
		delete(m.clients, userID)
		log.Printf("SSE client removed for user %d, total clients: %d", userID, len(m.clients))
	}
}

// Broadcast sends a message to all connected clients
func (m *SSEManager) Broadcast(message models.SSEMessage) {
	m.broadcast <- models.BroadcastMessage{
		UserID:  0, // 0 means broadcast to all
		Message: message,
	}
}

// SendToUser sends a message to a specific user
func (m *SSEManager) SendToUser(userID int, message models.SSEMessage) {
	m.broadcast <- models.BroadcastMessage{
		UserID:  userID,
		Message: message,
	}
}

// GetClientCount returns the number of connected clients
func (m *SSEManager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// GetConnectedUsers returns a list of connected user IDs
func (m *SSEManager) GetConnectedUsers() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]int, 0, len(m.clients))
	for userID := range m.clients {
		users = append(users, userID)
	}
	return users
}

// broadcastWorker processes broadcast messages
func (m *SSEManager) broadcastWorker() {
	for {
		select {
		case <-m.ctx.Done():
			log.Println("SSE broadcast worker stopped")
			return
		case msg := <-m.broadcast:
			m.sendMessage(msg)
		}
	}
}

// sendMessage sends a message to the appropriate clients
func (m *SSEManager) sendMessage(broadcastMsg models.BroadcastMessage) {
	// Convert message to SSE format
	jsonData, err := json.Marshal(broadcastMsg.Message)
	if err != nil {
		log.Printf("Error marshaling SSE message: %v", err)
		return
	}

	sseData := fmt.Sprintf("data: %s\n\n", string(jsonData))

	m.mu.RLock()
	defer m.mu.RUnlock()

	if broadcastMsg.UserID == 0 {
		// Broadcast to all clients
		for userID, clientChan := range m.clients {
			select {
			case clientChan <- sseData:
				// Message sent successfully
			case <-time.After(5 * time.Second):
				log.Printf("Timeout sending message to user %d", userID)
				// Remove the client if it's not responding
				go m.RemoveClient(userID)
			default:
				log.Printf("Client channel full for user %d", userID)
			}
		}
	} else {
		// Send to specific user
		if clientChan, exists := m.clients[broadcastMsg.UserID]; exists {
			select {
			case clientChan <- sseData:
				// Message sent successfully
			case <-time.After(5 * time.Second):
				log.Printf("Timeout sending message to user %d", broadcastMsg.UserID)
				go m.RemoveClient(broadcastMsg.UserID)
			default:
				log.Printf("Client channel full for user %d", broadcastMsg.UserID)
			}
		}
	}
}

// heartbeatWorker sends periodic heartbeat messages
func (m *SSEManager) heartbeatWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			log.Println("SSE heartbeat worker stopped")
			return
		case <-ticker.C:
			// Send heartbeat to all clients
			heartbeat := models.SSEMessage{
				Type: "heartbeat",
				Data: map[string]interface{}{
					"timestamp": time.Now().Unix(),
					"clients":   m.GetClientCount(),
				},
			}
			m.Broadcast(heartbeat)
		}
	}
}

// Close shuts down the SSE manager
func (m *SSEManager) Close() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all client channels
	for userID, clientChan := range m.clients {
		close(clientChan)
		delete(m.clients, userID)
	}

	log.Println("SSE manager closed")
}
