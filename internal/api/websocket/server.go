package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, you should check r.Origin
		return true
	},
}

// Server handles WebSocket connections
type Server struct {
	hub    *Hub
	logger *logrus.Logger
}

// NewServer creates a new WebSocket server
func NewServer(logger *logrus.Logger) *Server {
	hub := NewHub(logger)
	return &Server{
		hub:    hub,
		logger: logger,
	}
}

// Start starts the WebSocket server (runs the hub)
func (s *Server) Start() {
	go s.hub.Run()
}

// Stop stops the WebSocket server
func (s *Server) Stop() {
	s.hub.Stop()
}

// HandleWebSocket handles WebSocket connection requests
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := NewClient(s.hub, conn, s.logger)

	// Register client
	s.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	s.logger.Infof("WebSocket client connected from %s", r.RemoteAddr)
}

// GetHub returns the hub (for broadcasting events)
func (s *Server) GetHub() *Hub {
	return s.hub
}
