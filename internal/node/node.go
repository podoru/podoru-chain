package node

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/podoru/podoru-chain/internal/api/websocket"
	"github.com/podoru/podoru-chain/internal/blockchain"
	"github.com/podoru/podoru-chain/internal/consensus"
	"github.com/podoru/podoru-chain/internal/crypto"
	"github.com/podoru/podoru-chain/internal/network"
	"github.com/podoru/podoru-chain/internal/storage"
	"github.com/sirupsen/logrus"
)

// Node represents a blockchain node
type Node struct {
	config     *Config
	logger     *logrus.Logger
	storage    *storage.BadgerStore
	chain      *blockchain.Chain
	consensus  *consensus.PoAEngine
	p2pServer  *network.P2PServer
	mempool    *network.Mempool
	syncer     *network.Syncer
	privateKey *ecdsa.PrivateKey
	wsHub      *websocket.Hub
	stopChan   chan struct{}
}

// NewNode creates a new blockchain node
func NewNode(config *Config) (*Node, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	node := &Node{
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
	}

	// Load private key if this is a producer node
	if config.IsProducer() {
		privateKey, err := crypto.LoadPrivateKeyFromFile(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		node.privateKey = privateKey

		// Verify address matches
		derivedAddr, err := crypto.AddressFromPrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to derive address: %w", err)
		}
		if crypto.NormalizeAddress(derivedAddr) != crypto.NormalizeAddress(config.Address) {
			return nil, fmt.Errorf("address mismatch: config=%s, derived=%s", config.Address, derivedAddr)
		}
	}

	return node, nil
}

// Start starts the node
func (n *Node) Start() error {
	n.logger.Infof("Starting Podoru Chain node (type: %s)...", n.config.NodeType)

	// Initialize storage
	n.logger.Info("Initializing storage...")
	store, err := storage.NewBadgerStore(n.config.DataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	n.storage = store

	// Initialize consensus
	n.logger.Info("Initializing consensus engine...")
	consensusEngine, err := consensus.NewPoAEngine(n.config.Authorities, n.config.BlockTime)
	if err != nil {
		return fmt.Errorf("failed to initialize consensus: %w", err)
	}
	n.consensus = consensusEngine

	// Initialize blockchain
	n.logger.Info("Initializing blockchain...")
	n.chain = blockchain.NewChain(n.storage, n.config.Authorities)

	// Try to load existing chain or create genesis
	if err := n.initializeChain(); err != nil {
		return fmt.Errorf("failed to initialize chain: %w", err)
	}

	// Initialize mempool
	n.logger.Info("Initializing mempool...")
	n.mempool = network.NewMempool()

	// Initialize P2P server
	n.logger.Info("Initializing P2P network...")
	n.p2pServer = network.NewP2PServer(n.config.P2PBindAddr, n.config.P2PPort, n.logger)
	n.registerP2PHandlers()

	if err := n.p2pServer.Start(); err != nil {
		return fmt.Errorf("failed to start P2P server: %w", err)
	}

	// Connect to bootstrap peers
	n.logger.Info("Connecting to bootstrap peers...")
	for _, peer := range n.config.BootstrapPeers {
		if err := n.p2pServer.ConnectToPeer(peer); err != nil {
			n.logger.Warnf("Failed to connect to bootstrap peer %s: %v", peer, err)
		}
	}

	// Initialize syncer
	n.logger.Info("Initializing syncer...")
	n.syncer = network.NewSyncer(n.chain, n.p2pServer, n.mempool, n.logger)

	// Start auto-sync to catch up with peers
	n.logger.Info("Starting auto-sync...")
	n.syncer.StartAutoSync()

	// Start block production if this is a producer node
	if n.config.IsProducer() {
		n.logger.Info("Starting block production...")
		go n.blockProductionLoop()
	}

	n.logger.Info("Node started successfully")
	return nil
}

// initializeChain initializes the blockchain (load or create genesis)
func (n *Node) initializeChain() error {
	// Try to load existing chain
	if err := n.chain.LoadFromStorage(); err != nil {
		// Chain doesn't exist, create genesis
		n.logger.Info("Creating genesis block...")

		genesisConfig, err := blockchain.LoadGenesisConfig(n.config.GenesisPath)
		if err != nil {
			return fmt.Errorf("failed to load genesis config: %w", err)
		}

		genesisBlock := blockchain.CreateGenesisBlock(genesisConfig)

		if err := n.chain.Initialize(genesisBlock); err != nil {
			return fmt.Errorf("failed to initialize chain with genesis: %w", err)
		}

		n.logger.Info("Genesis block created")
	} else {
		n.logger.Infof("Loaded blockchain from storage (height: %d)", n.chain.GetHeight())
	}

	return nil
}

// registerP2PHandlers registers message handlers for P2P network
func (n *Node) registerP2PHandlers() {
	// Handle new block messages
	n.p2pServer.RegisterHandler(network.MsgTypeNewBlock, n.handleNewBlock)

	// Handle new transaction messages
	n.p2pServer.RegisterHandler(network.MsgTypeNewTransaction, n.handleNewTransaction)

	// Handle get blocks messages
	n.p2pServer.RegisterHandler(network.MsgTypeGetBlocks, n.handleGetBlocks)

	// Handle get height messages
	n.p2pServer.RegisterHandler(network.MsgTypeGetHeight, n.handleGetHeight)

	// Handle ping messages
	n.p2pServer.RegisterHandler(network.MsgTypePing, n.handlePing)
}

// handleNewBlock handles incoming new block messages
func (n *Node) handleNewBlock(peer *network.Peer, msg *network.Message) error {
	n.logger.Info("Received new block from peer")

	// Convert payload to correct type (JSON unmarshaling creates map[string]interface{})
	var newBlockMsg network.NewBlockMessage
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &newBlockMsg); err != nil {
		return fmt.Errorf("failed to unmarshal new block message: %w", err)
	}

	block := newBlockMsg.Block
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	currentBlock := n.chain.GetCurrentBlock()
	currentHeight := currentBlock.Header.Height

	// Check if block is already processed (stale)
	if block.Header.Height <= currentHeight {
		n.logger.Debugf("Ignoring block at height %d (current: %d)", block.Header.Height, currentHeight)
		return nil
	}

	// Check if block is the NEXT expected block
	expectedHeight := currentHeight + 1
	if block.Header.Height == expectedHeight {
		// This is the next block - add it normally
		if err := n.chain.AddBlock(block); err != nil {
			n.logger.Errorf("Failed to add received block: %v", err)
			return err
		}
		n.logger.Infof("Added block %d from peer (txs: %d)", block.Header.Height, len(block.Transactions))
		n.mempool.RemoveTransactions(block.Transactions)

		// Broadcast block event via WebSocket
		n.broadcastBlockEvent(block)

		return nil
	}

	// Block is too far ahead - trigger sync instead of rejecting
	if block.Header.Height > expectedHeight {
		n.logger.Warnf("Block %d is ahead of current height %d, triggering sync...",
			block.Header.Height, currentHeight)

		// Trigger sync in background (non-blocking)
		n.syncer.TriggerSync()

		// Don't return error - this is expected behavior for catching up
		return nil
	}

	return nil
}

// handleNewTransaction handles incoming new transaction messages
func (n *Node) handleNewTransaction(peer *network.Peer, msg *network.Message) error {
	n.logger.Info("Received new transaction from peer")

	// Convert payload to correct type (JSON unmarshaling creates map[string]interface{})
	var newTxMsg network.NewTransactionMessage
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &newTxMsg); err != nil {
		return fmt.Errorf("failed to unmarshal new transaction message: %w", err)
	}

	tx := newTxMsg.Transaction
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// Add transaction to mempool (this will validate it)
	if err := n.mempool.AddTransaction(tx); err != nil {
		n.logger.Debugf("Failed to add transaction to mempool: %v", err)
		return nil // Don't return error for duplicate/invalid txs
	}

	n.logger.Infof("Added transaction %x to mempool", tx.ID)

	// Broadcast transaction event via WebSocket
	n.broadcastTransactionEvent(tx, "pending")

	return nil
}

// handleGetBlocks handles get blocks requests
func (n *Node) handleGetBlocks(peer *network.Peer, msg *network.Message) error {
	n.logger.Info("Received get blocks request from peer")

	// Parse request
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return err
	}

	var req network.GetBlocksMessage
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		return err
	}

	// Retrieve blocks
	blocks := make([]*blockchain.Block, 0, req.ToHeight-req.FromHeight+1)
	for h := req.FromHeight; h <= req.ToHeight; h++ {
		block, err := n.chain.GetBlockByHeight(h)
		if err != nil {
			break // No more blocks
		}
		blocks = append(blocks, block)
	}

	n.logger.Infof("Sending %d blocks (height %d to %d) to peer %s", len(blocks), req.FromHeight, req.ToHeight, peer.ID)

	// Send response
	response := &network.Message{
		Type:    network.MsgTypeBlocks,
		Payload: &network.BlocksMessage{Blocks: blocks},
	}

	return n.p2pServer.SendMessage(peer, response)
}

// handleGetHeight handles get height requests
func (n *Node) handleGetHeight(peer *network.Peer, msg *network.Message) error {
	height := n.chain.GetHeight()

	response := &network.Message{
		Type:    network.MsgTypeHeight,
		Payload: &network.HeightMessage{Height: height},
	}

	n.logger.Debugf("Responding to height request from %s: height=%d", peer.ID, height)
	return n.p2pServer.SendMessage(peer, response)
}

// handlePing handles ping messages
func (n *Node) handlePing(peer *network.Peer, msg *network.Message) error {
	// Send pong response
	pong := &network.Message{
		Type:    network.MsgTypePong,
		Payload: &network.PongMessage{Timestamp: time.Now().Unix()},
	}
	return n.p2pServer.SendMessage(peer, pong)
}

// blockProductionLoop runs the block production loop for producer nodes
func (n *Node) blockProductionLoop() {
	ticker := time.NewTicker(n.config.BlockTime)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			if err := n.produceBlock(); err != nil {
				n.logger.Errorf("Failed to produce block: %v", err)
			}
		}
	}
}

// produceBlock produces a new block
func (n *Node) produceBlock() error {
	currentBlock := n.chain.GetCurrentBlock()
	nextHeight := currentBlock.Header.Height + 1

	// Check if it's our turn to produce
	if !n.consensus.CanProduceBlock(nextHeight, n.config.Address) {
		return nil // Not our turn
	}

	// Check if enough time has passed
	if !n.consensus.ShouldProduceBlock(currentBlock.Header.Timestamp) {
		return nil // Too soon
	}

	n.logger.Infof("Producing block at height %d...", nextHeight)

	// Get pending transactions from mempool
	transactions := n.mempool.GetPendingTransactions(blockchain.MaxTransactionsPerBlock)

	// Calculate merkle root
	merkleRoot := blockchain.CalculateMerkleRoot(transactions)

	// Calculate state root AFTER applying transactions
	stateRoot, err := n.chain.CalculateStateRootWithTransactions(transactions)
	if err != nil {
		return fmt.Errorf("failed to calculate state root: %w", err)
	}

	// Create block header
	header := &blockchain.BlockHeader{
		Version:      1,
		Height:       nextHeight,
		PreviousHash: currentBlock.Hash(),
		Timestamp:    time.Now().Unix(),
		MerkleRoot:   merkleRoot,
		StateRoot:    stateRoot,
		ProducerAddr: n.config.Address,
		Nonce:        0,
	}

	// Create block
	block := blockchain.NewBlock(header, transactions)

	// Sign block
	if err := block.Sign(n.privateKey); err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	// Add block to chain
	if err := n.chain.AddBlock(block); err != nil {
		return fmt.Errorf("failed to add block to chain: %w", err)
	}

	// Remove transactions from mempool
	n.mempool.RemoveTransactions(transactions)

	// Broadcast block to peers
	msg := &network.Message{
		Type:    network.MsgTypeNewBlock,
		Payload: &network.NewBlockMessage{Block: block},
	}
	n.p2pServer.BroadcastMessage(msg)

	// Broadcast block event via WebSocket
	n.broadcastBlockEvent(block)

	n.logger.Infof("Block %d produced successfully (txs: %d)", nextHeight, len(transactions))

	return nil
}

// SubmitTransaction submits a transaction to the mempool
func (n *Node) SubmitTransaction(tx *blockchain.Transaction) error {
	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Add to mempool
	if err := n.mempool.AddTransaction(tx); err != nil {
		return fmt.Errorf("failed to add to mempool: %w", err)
	}

	// Broadcast to peers
	msg := &network.Message{
		Type:    network.MsgTypeNewTransaction,
		Payload: &network.NewTransactionMessage{Transaction: tx},
	}
	n.p2pServer.BroadcastMessage(msg)

	// Broadcast transaction event via WebSocket
	n.broadcastTransactionEvent(tx, "pending")

	return nil
}

// GetChain returns the blockchain
func (n *Node) GetChain() *blockchain.Chain {
	return n.chain
}

// GetMempool returns the mempool
func (n *Node) GetMempool() *network.Mempool {
	return n.mempool
}

// GetP2PServer returns the P2P server
func (n *Node) GetP2PServer() *network.P2PServer {
	return n.p2pServer
}

// SetWebSocketHub sets the WebSocket hub for broadcasting events
func (n *Node) SetWebSocketHub(hub *websocket.Hub) {
	n.wsHub = hub
}

// broadcastBlockEvent broadcasts a new block event via WebSocket
func (n *Node) broadcastBlockEvent(block *blockchain.Block) {
	if n.wsHub != nil {
		event := websocket.NewBlockEvent(block)
		n.wsHub.Broadcast(event)
	}
}

// broadcastTransactionEvent broadcasts a new transaction event via WebSocket
func (n *Node) broadcastTransactionEvent(tx *blockchain.Transaction, status string) {
	if n.wsHub != nil {
		event := websocket.NewTransactionEvent(tx, status)
		n.wsHub.Broadcast(event)
	}
}

// Stop stops the node
func (n *Node) Stop() error {
	n.logger.Info("Stopping node...")

	close(n.stopChan)

	// Stop P2P server
	if n.p2pServer != nil {
		n.p2pServer.Stop()
	}

	// Close storage
	if n.storage != nil {
		if err := n.storage.Close(); err != nil {
			return fmt.Errorf("failed to close storage: %w", err)
		}
	}

	n.logger.Info("Node stopped")
	return nil
}
