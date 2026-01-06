package network

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Peer represents a connected peer
type Peer struct {
	ID      string
	Conn    net.Conn
	Address string
	writer  *bufio.Writer
	mu      sync.Mutex
}

// P2PServer manages peer-to-peer connections
type P2PServer struct {
	mu              sync.RWMutex
	bindAddr        string
	port            int
	peers           map[string]*Peer
	listener        net.Listener
	messageHandlers map[MessageType]MessageHandler
	logger          *logrus.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// MessageHandler is a function that handles incoming messages
type MessageHandler func(peer *Peer, msg *Message) error

// NewP2PServer creates a new P2P server
func NewP2PServer(bindAddr string, port int, logger *logrus.Logger) *P2PServer {
	if logger == nil {
		logger = logrus.New()
	}

	return &P2PServer{
		bindAddr:        bindAddr,
		port:            port,
		peers:           make(map[string]*Peer),
		messageHandlers: make(map[MessageType]MessageHandler),
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

// RegisterHandler registers a message handler for a specific message type
func (p2p *P2PServer) RegisterHandler(msgType MessageType, handler MessageHandler) {
	p2p.mu.Lock()
	defer p2p.mu.Unlock()

	p2p.messageHandlers[msgType] = handler
}

// Start starts the P2P server
func (p2p *P2PServer) Start() error {
	addr := fmt.Sprintf("%s:%d", p2p.bindAddr, p2p.port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start P2P server: %w", err)
	}

	p2p.listener = listener
	p2p.logger.Infof("P2P server listening on %s", addr)

	p2p.wg.Add(1)
	go p2p.acceptLoop()

	return nil
}

// acceptLoop accepts incoming connections
func (p2p *P2PServer) acceptLoop() {
	defer p2p.wg.Done()

	for {
		select {
		case <-p2p.stopChan:
			return
		default:
		}

		p2p.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
		conn, err := p2p.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			p2p.logger.Errorf("Error accepting connection: %v", err)
			continue
		}

		p2p.wg.Add(1)
		go p2p.handlePeer(conn)
	}
}

// handlePeer handles communication with a peer
func (p2p *P2PServer) handlePeer(conn net.Conn) {
	defer p2p.wg.Done()
	defer conn.Close()

	peer := &Peer{
		ID:      conn.RemoteAddr().String(),
		Conn:    conn,
		Address: conn.RemoteAddr().String(),
		writer:  bufio.NewWriter(conn),
	}

	// Add peer
	p2p.addPeer(peer)
	defer p2p.removePeer(peer.ID)

	p2p.logger.Infof("New peer connected: %s", peer.ID)

	// Read messages
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-p2p.stopChan:
			return
		default:
		}

		msg, err := p2p.readMessage(reader)
		if err != nil {
			if err != io.EOF {
				p2p.logger.Errorf("Error reading message from %s: %v", peer.ID, err)
			}
			return
		}

		// Handle message
		if err := p2p.handleMessage(peer, msg); err != nil {
			p2p.logger.Errorf("Error handling message from %s: %v", peer.ID, err)
		}
	}
}

// readMessage reads a message from a reader (length-prefixed JSON)
func (p2p *P2PServer) readMessage(reader *bufio.Reader) (*Message, error) {
	// Read message length (4 bytes)
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Prevent DOS attacks
	if length > 10*1024*1024 { // 10 MB max
		return nil, errors.New("message too large")
	}

	// Read message data
	msgBytes := make([]byte, length)
	if _, err := io.ReadFull(reader, msgBytes); err != nil {
		return nil, err
	}

	// Unmarshal message
	var msg Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// SendMessage sends a message to a peer
func (p2p *P2PServer) SendMessage(peer *Peer, msg *Message) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// Marshal message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write length prefix
	length := uint32(len(msgBytes))
	if err := binary.Write(peer.writer, binary.BigEndian, length); err != nil {
		return err
	}

	// Write message
	if _, err := peer.writer.Write(msgBytes); err != nil {
		return err
	}

	return peer.writer.Flush()
}

// BroadcastMessage broadcasts a message to all peers
func (p2p *P2PServer) BroadcastMessage(msg *Message) {
	p2p.mu.RLock()
	peers := make([]*Peer, 0, len(p2p.peers))
	for _, peer := range p2p.peers {
		peers = append(peers, peer)
	}
	p2p.mu.RUnlock()

	for _, peer := range peers {
		if err := p2p.SendMessage(peer, msg); err != nil {
			p2p.logger.Errorf("Failed to send message to %s: %v", peer.ID, err)
		}
	}
}

// handleMessage handles an incoming message
func (p2p *P2PServer) handleMessage(peer *Peer, msg *Message) error {
	p2p.mu.RLock()
	handler, exists := p2p.messageHandlers[msg.Type]
	p2p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler for message type %d", msg.Type)
	}

	return handler(peer, msg)
}

// ConnectToPeer connects to a remote peer
func (p2p *P2PServer) ConnectToPeer(address string) error {
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	p2p.wg.Add(1)
	go p2p.handlePeer(conn)

	return nil
}

// addPeer adds a peer to the peer list
func (p2p *P2PServer) addPeer(peer *Peer) {
	p2p.mu.Lock()
	defer p2p.mu.Unlock()

	p2p.peers[peer.ID] = peer
}

// removePeer removes a peer from the peer list
func (p2p *P2PServer) removePeer(peerID string) {
	p2p.mu.Lock()
	defer p2p.mu.Unlock()

	delete(p2p.peers, peerID)
	p2p.logger.Infof("Peer disconnected: %s", peerID)
}

// GetPeers returns a list of connected peers
func (p2p *P2PServer) GetPeers() []*Peer {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	peers := make([]*Peer, 0, len(p2p.peers))
	for _, peer := range p2p.peers {
		peers = append(peers, peer)
	}

	return peers
}

// PeerCount returns the number of connected peers
func (p2p *P2PServer) PeerCount() int {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	return len(p2p.peers)
}

// Stop stops the P2P server
func (p2p *P2PServer) Stop() {
	close(p2p.stopChan)

	if p2p.listener != nil {
		p2p.listener.Close()
	}

	// Close all peer connections
	p2p.mu.Lock()
	for _, peer := range p2p.peers {
		peer.Conn.Close()
	}
	p2p.mu.Unlock()

	p2p.wg.Wait()
	p2p.logger.Info("P2P server stopped")
}
