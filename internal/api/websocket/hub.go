package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *Event

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	logger *logrus.Logger

	// Stop channel
	stopChan chan struct{}
}

// NewHub creates a new Hub
func NewHub(logger *logrus.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	h.logger.Info("WebSocket hub started")

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debugf("Client connected (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debugf("Client disconnected (total: %d)", len(h.clients))

		case event := <-h.broadcast:
			h.broadcastEvent(event)

		case <-h.stopChan:
			h.logger.Info("WebSocket hub stopping")
			h.closeAllClients()
			return
		}
	}
}

// broadcastEvent sends an event to all subscribed clients
func (h *Hub) broadcastEvent(event *Event) {
	// Set timestamp if not already set
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Marshal event to JSON once
	message, err := json.Marshal(event)
	if err != nil {
		h.logger.Errorf("Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Send to all subscribed clients
	for client := range h.clients {
		if client.isSubscribed(event.Type) {
			select {
			case client.send <- message:
				// Message sent successfully
			default:
				// Client's send buffer is full, close the connection
				h.logger.Warnf("Client buffer full, closing connection")
				go func(c *Client) {
					h.unregister <- c
					c.conn.Close()
				}(client)
			}
		}
	}
}

// closeAllClients closes all client connections
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}
}

// Stop stops the hub
func (h *Hub) Stop() {
	close(h.stopChan)
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event *Event) {
	select {
	case h.broadcast <- event:
	default:
		h.logger.Warn("Broadcast channel full, dropping event")
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
