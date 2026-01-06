package blockchain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Storage interface for blockchain data persistence
type Storage interface {
	SaveBlock(block *Block) error
	GetBlock(hash []byte) (*Block, error)
	GetBlockByHeight(height uint64) (*Block, error)
	SaveTransaction(tx *Transaction) error
	GetTransaction(hash []byte) (*Transaction, error)
	SaveState(key string, value []byte) error
	GetState(key string) ([]byte, error)
	DeleteState(key string) error
	GetLatestBlockHeight() (uint64, error)
	SaveBlockHeight(height uint64) error
	ScanStateByPrefix(prefix string, limit int) (map[string][]byte, error)
	GetAllStateKeys(limit int) ([]string, error)
	Close() error
}

// State represents the current key-value state
type State struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewState creates a new state
func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

// Set sets a key-value pair
func (s *State) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get gets a value by key
func (s *State) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Delete deletes a key
func (s *State) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// CalculateRoot calculates the merkle root of the state
func (s *State) CalculateRoot() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.data) == 0 {
		return make([]byte, 32)
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create merkle tree of state entries
	hashes := make([][]byte, len(keys))
	for i, k := range keys {
		entry := append([]byte(k), s.data[k]...)
		hash := sha256.Sum256(entry)
		hashes[i] = hash[:]
	}

	return buildMerkleTree(hashes)
}

// Clone creates a deep copy of the state
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewState()
	for k, v := range s.data {
		newState.data[k] = append([]byte{}, v...)
	}
	return newState
}

// Chain manages the blockchain
type Chain struct {
	mu           sync.RWMutex
	storage      Storage
	currentBlock *Block
	height       uint64
	state        *State
	authorities  []string
	nonces       map[string]uint64 // Track nonces per address
}

// NewChain creates a new blockchain
func NewChain(storage Storage, authorities []string) *Chain {
	return &Chain{
		storage:     storage,
		state:       NewState(),
		authorities: authorities,
		nonces:      make(map[string]uint64),
	}
}

// Initialize initializes the chain with a genesis block
func (c *Chain) Initialize(genesisBlock *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if chain is already initialized
	height, err := c.storage.GetLatestBlockHeight()
	if err == nil && height > 0 {
		return errors.New("chain already initialized")
	}

	// Validate genesis block
	if err := validateGenesisBlock(genesisBlock); err != nil {
		return fmt.Errorf("invalid genesis block: %w", err)
	}

	// Apply genesis transactions to state
	if err := c.applyTransactions(genesisBlock.Transactions); err != nil {
		return fmt.Errorf("failed to apply genesis transactions: %w", err)
	}

	// Update state root in genesis block
	genesisBlock.Header.StateRoot = c.state.CalculateRoot()

	// Save genesis block
	if err := c.storage.SaveBlock(genesisBlock); err != nil {
		return fmt.Errorf("failed to save genesis block: %w", err)
	}

	// Save transactions
	for _, tx := range genesisBlock.Transactions {
		if err := c.storage.SaveTransaction(tx); err != nil {
			return fmt.Errorf("failed to save genesis transaction: %w", err)
		}
	}

	// Update chain state
	c.currentBlock = genesisBlock
	c.height = 0

	if err := c.storage.SaveBlockHeight(0); err != nil {
		return fmt.Errorf("failed to save block height: %w", err)
	}

	return nil
}

// LoadFromStorage loads the blockchain state from storage
func (c *Chain) LoadFromStorage() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get latest height
	height, err := c.storage.GetLatestBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get latest height: %w", err)
	}

	// Load latest block
	block, err := c.storage.GetBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("failed to load block at height %d: %w", height, err)
	}

	c.currentBlock = block
	c.height = height

	// Rebuild state from genesis to current height
	// For now, we'll need to replay all blocks
	// In a production system, you'd want to store state snapshots
	return c.rebuildState()
}

// rebuildState rebuilds the state by replaying all blocks
func (c *Chain) rebuildState() error {
	c.state = NewState()
	c.nonces = make(map[string]uint64)

	// Replay all blocks from genesis to current height
	for h := uint64(0); h <= c.height; h++ {
		block, err := c.storage.GetBlockByHeight(h)
		if err != nil {
			return fmt.Errorf("failed to load block at height %d: %w", h, err)
		}

		if err := c.applyTransactions(block.Transactions); err != nil {
			return fmt.Errorf("failed to apply transactions at height %d: %w", h, err)
		}
	}

	return nil
}

// AddBlock adds a validated block to the chain
func (c *Chain) AddBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate block
	if err := ValidateBlock(block, c.currentBlock, c.authorities); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Validate state root by applying transactions to a temporary state
	tempState := c.state.Clone()
	if err := c.applyTransactionsToState(tempState, block.Transactions); err != nil {
		return fmt.Errorf("failed to apply transactions: %w", err)
	}

	calculatedStateRoot := tempState.CalculateRoot()
	if !bytes.Equal(calculatedStateRoot, block.Header.StateRoot) {
		return errors.New("invalid state root")
	}

	// Apply transactions to actual state
	if err := c.applyTransactions(block.Transactions); err != nil {
		return fmt.Errorf("failed to apply transactions: %w", err)
	}

	// Save block and transactions
	if err := c.storage.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	for _, tx := range block.Transactions {
		if err := c.storage.SaveTransaction(tx); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}
	}

	// Update chain state
	c.currentBlock = block
	c.height = block.Header.Height

	if err := c.storage.SaveBlockHeight(c.height); err != nil {
		return fmt.Errorf("failed to save block height: %w", err)
	}

	return nil
}

// applyTransactions applies transactions to the current state
func (c *Chain) applyTransactions(transactions []*Transaction) error {
	return c.applyTransactionsToState(c.state, transactions)
}

// applyTransactionsToState applies transactions to a given state
func (c *Chain) applyTransactionsToState(state *State, transactions []*Transaction) error {
	for _, tx := range transactions {
		for _, op := range tx.Data.Operations {
			switch op.Type {
			case OpTypeSet:
				state.Set(op.Key, op.Value)
				// Also persist to storage
				if state == c.state {
					if err := c.storage.SaveState(op.Key, op.Value); err != nil {
						return fmt.Errorf("failed to save state: %w", err)
					}
				}
			case OpTypeDelete:
				state.Delete(op.Key)
				// Also delete from storage
				if state == c.state {
					if err := c.storage.DeleteState(op.Key); err != nil {
						return fmt.Errorf("failed to delete state: %w", err)
					}
				}
			default:
				return fmt.Errorf("unknown operation type: %s", op.Type)
			}
		}

		// Update nonce
		if state == c.state && tx.From != "0x0000000000000000000000000000000000000000" {
			c.nonces[tx.From] = tx.Nonce + 1
		}
	}

	return nil
}

// GetState retrieves a value from the current state
func (c *Chain) GetState(key string) ([]byte, error) {
	value, exists := c.state.Get(key)
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

// GetCurrentBlock returns the current block
func (c *Chain) GetCurrentBlock() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentBlock
}

// GetHeight returns the current chain height
func (c *Chain) GetHeight() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// GetBlockByHeight retrieves a block by height
func (c *Chain) GetBlockByHeight(height uint64) (*Block, error) {
	return c.storage.GetBlockByHeight(height)
}

// GetBlockByHash retrieves a block by hash
func (c *Chain) GetBlockByHash(hash []byte) (*Block, error) {
	return c.storage.GetBlock(hash)
}

// GetTransaction retrieves a transaction by hash
func (c *Chain) GetTransaction(hash []byte) (*Transaction, error) {
	return c.storage.GetTransaction(hash)
}

// GetNonce returns the next nonce for an address
func (c *Chain) GetNonce(address string) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nonce, exists := c.nonces[address]
	if !exists {
		return 0
	}
	return nonce
}

// GetStateRoot returns the current state root
func (c *Chain) GetStateRoot() []byte {
	return c.state.CalculateRoot()
}

// QueryStateByPrefix queries all state keys with a given prefix
func (c *Chain) QueryStateByPrefix(prefix string, limit int) (map[string][]byte, error) {
	return c.storage.ScanStateByPrefix(prefix, limit)
}

// GetAllStateKeys returns all state keys
func (c *Chain) GetAllStateKeys(limit int) ([]string, error) {
	return c.storage.GetAllStateKeys(limit)
}

// GetAuthorities returns the list of authorities
func (c *Chain) GetAuthorities() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string{}, c.authorities...)
}

// ChainInfo contains information about the chain
type ChainInfo struct {
	Height       uint64 `json:"height"`
	CurrentHash  string `json:"current_hash"`
	GenesisHash  string `json:"genesis_hash"`
	Authorities  []string `json:"authorities"`
	StateRoot    string `json:"state_root"`
}

// GetChainInfo returns information about the chain
func (c *Chain) GetChainInfo() (*ChainInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentBlock == nil {
		return nil, errors.New("chain not initialized")
	}

	genesisBlock, err := c.storage.GetBlockByHeight(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get genesis block: %w", err)
	}

	return &ChainInfo{
		Height:      c.height,
		CurrentHash: fmt.Sprintf("0x%x", c.currentBlock.Hash()),
		GenesisHash: fmt.Sprintf("0x%x", genesisBlock.Hash()),
		Authorities: c.authorities,
		StateRoot:   fmt.Sprintf("0x%x", c.GetStateRoot()),
	}, nil
}
