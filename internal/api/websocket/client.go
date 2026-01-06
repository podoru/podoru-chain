package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a WebSocket client connection
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Subscribed event types
	subscriptions map[EventType]bool

	logger *logrus.Logger
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, logger *logrus.Logger) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[EventType]bool),
		logger:        logger,
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Parse subscription message
		var subMsg SubscribeMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			c.logger.Warnf("Failed to parse subscription message: %v", err)
			continue
		}

		// Handle subscription/unsubscription
		c.handleSubscription(&subMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscription processes subscription/unsubscription requests
func (c *Client) handleSubscription(msg *SubscribeMessage) {
	switch msg.Action {
	case "subscribe":
		for _, eventType := range msg.Events {
			c.subscriptions[eventType] = true
			c.logger.Debugf("Client subscribed to %s", eventType)
		}
	case "unsubscribe":
		for _, eventType := range msg.Events {
			delete(c.subscriptions, eventType)
			c.logger.Debugf("Client unsubscribed from %s", eventType)
		}
	default:
		c.logger.Warnf("Unknown subscription action: %s", msg.Action)
	}
}

// isSubscribed checks if the client is subscribed to an event type
func (c *Client) isSubscribed(eventType EventType) bool {
	// If no subscriptions, send all events
	if len(c.subscriptions) == 0 {
		return true
	}
	return c.subscriptions[eventType]
}
