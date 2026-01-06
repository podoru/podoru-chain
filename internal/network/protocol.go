package network

import (
	"github.com/podoru/podoru-chain/internal/blockchain"
)

// MessageType defines different P2P message types
type MessageType uint8

const (
	MsgTypePing MessageType = iota
	MsgTypePong
	MsgTypeGetPeers
	MsgTypePeers
	MsgTypeNewBlock
	MsgTypeGetBlocks
	MsgTypeBlocks
	MsgTypeNewTransaction
	MsgTypeGetBlockByHeight
	MsgTypeGetBlockByHash
	MsgTypeGetState
	MsgTypeGetHeight
	MsgTypeHeight
)

// Message is the envelope for all P2P messages
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
	From    string      `json:"from"` // Sender peer ID
}

// PingMessage is sent to check if a peer is alive
type PingMessage struct {
	Timestamp int64 `json:"timestamp"`
}

// PongMessage is the response to a ping
type PongMessage struct {
	Timestamp int64 `json:"timestamp"`
}

// GetPeersMessage requests peer information
type GetPeersMessage struct{}

// PeerInfo contains information about a peer
type PeerInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// PeersMessage contains a list of peers
type PeersMessage struct {
	Peers []PeerInfo `json:"peers"`
}

// NewBlockMessage announces a new block
type NewBlockMessage struct {
	Block *blockchain.Block `json:"block"`
}

// GetBlocksMessage requests blocks in a range
type GetBlocksMessage struct {
	FromHeight uint64 `json:"from_height"`
	ToHeight   uint64 `json:"to_height"`
}

// BlocksMessage responds with blocks
type BlocksMessage struct {
	Blocks []*blockchain.Block `json:"blocks"`
}

// NewTransactionMessage broadcasts a new transaction
type NewTransactionMessage struct {
	Transaction *blockchain.Transaction `json:"transaction"`
}

// GetBlockByHeightMessage requests a specific block by height
type GetBlockByHeightMessage struct {
	Height uint64 `json:"height"`
}

// GetBlockByHashMessage requests a specific block by hash
type GetBlockByHashMessage struct {
	Hash []byte `json:"hash"`
}

// GetStateMessage requests a state value
type GetStateMessage struct {
	Key string `json:"key"`
}

// StateMessage responds with a state value
type StateMessage struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// GetHeightMessage requests the current chain height
type GetHeightMessage struct{}

// HeightMessage responds with the current height
type HeightMessage struct {
	Height uint64 `json:"height"`
}
