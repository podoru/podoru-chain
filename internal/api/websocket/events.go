package websocket

import (
	"github.com/podoru/podoru-chain/internal/blockchain"
)

// EventType defines the type of event being broadcast
type EventType string

const (
	EventNewBlock       EventType = "new_block"
	EventNewTransaction EventType = "new_transaction"
	EventChainUpdate    EventType = "chain_update"
	EventMempoolUpdate  EventType = "mempool_update"
)

// Event represents a WebSocket event message
type Event struct {
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// BlockEvent represents a new block event
type BlockEvent struct {
	Height           uint64 `json:"height"`
	Hash             string `json:"hash"`
	Timestamp        int64  `json:"timestamp"`
	TransactionCount int    `json:"transaction_count"`
	Producer         string `json:"producer"`
	PreviousHash     string `json:"previous_hash"`
}

// TransactionEvent represents a transaction event
type TransactionEvent struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"` // "pending" or "confirmed"
	Nonce     uint64 `json:"nonce"`
}

// ChainUpdateEvent represents a chain state update
type ChainUpdateEvent struct {
	Height      uint64   `json:"height"`
	CurrentHash string   `json:"current_hash"`
	Authorities []string `json:"authorities"`
}

// MempoolUpdateEvent represents mempool changes
type MempoolUpdateEvent struct {
	Count        int      `json:"count"`
	RecentHashes []string `json:"recent_hashes"`
}

// SubscribeMessage represents a subscription request from client
type SubscribeMessage struct {
	Action string      `json:"action"` // "subscribe" or "unsubscribe"
	Events []EventType `json:"events"`
}

// NewBlockEvent creates a block event from a blockchain block
func NewBlockEvent(block *blockchain.Block) *Event {
	return &Event{
		Type: EventNewBlock,
		Data: &BlockEvent{
			Height:           block.Header.Height,
			Hash:             block.HashString(),
			Timestamp:        block.Header.Timestamp,
			TransactionCount: len(block.Transactions),
			Producer:         block.Header.ProducerAddr,
			PreviousHash:     block.Header.PreviousHashString(),
		},
		Timestamp: block.Header.Timestamp,
	}
}

// NewTransactionEvent creates a transaction event
func NewTransactionEvent(tx *blockchain.Transaction, status string) *Event {
	return &Event{
		Type: EventNewTransaction,
		Data: &TransactionEvent{
			Hash:      tx.HashString(),
			From:      tx.From,
			Timestamp: tx.Timestamp,
			Status:    status,
			Nonce:     tx.Nonce,
		},
		Timestamp: tx.Timestamp,
	}
}

// NewChainUpdateEvent creates a chain update event
func NewChainUpdateEvent(height uint64, hash string, authorities []string) *Event {
	return &Event{
		Type: EventChainUpdate,
		Data: &ChainUpdateEvent{
			Height:      height,
			CurrentHash: hash,
			Authorities: authorities,
		},
		Timestamp: 0, // Will be set by hub
	}
}

// NewMempoolUpdateEvent creates a mempool update event
func NewMempoolUpdateEvent(count int, recentHashes []string) *Event {
	return &Event{
		Type: EventMempoolUpdate,
		Data: &MempoolUpdateEvent{
			Count:        count,
			RecentHashes: recentHashes,
		},
		Timestamp: 0, // Will be set by hub
	}
}
