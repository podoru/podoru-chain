package blockchain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
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
	gasConfig    *GasConfig        // Gas fee configuration (nil for legacy chains)
	tokenConfig  *TokenConfig      // Token configuration (nil for legacy chains)
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

// NewChainWithConfig creates a new blockchain with gas and token configuration
func NewChainWithConfig(storage Storage, authorities []string, gasConfig *GasConfig, tokenConfig *TokenConfig) *Chain {
	return &Chain{
		storage:     storage,
		state:       NewState(),
		authorities: authorities,
		nonces:      make(map[string]uint64),
		gasConfig:   gasConfig,
		tokenConfig: tokenConfig,
	}
}

// SetGasConfig sets the gas configuration
func (c *Chain) SetGasConfig(config *GasConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gasConfig = config
}

// GetGasConfig returns the gas configuration
func (c *Chain) GetGasConfig() *GasConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gasConfig
}

// SetTokenConfig sets the token configuration
func (c *Chain) SetTokenConfig(config *TokenConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenConfig = config
}

// GetTokenConfig returns the token configuration
func (c *Chain) GetTokenConfig() *TokenConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tokenConfig
}

// HasGasFees returns true if gas fees are enabled
func (c *Chain) HasGasFees() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gasConfig != nil && !c.gasConfig.IsZeroFee()
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
			case OpTypeMint:
				// MINT operation: add amount to existing balance
				if err := c.applyMintOperation(state, op); err != nil {
					return err
				}
			case OpTypeTransfer:
				// TRANSFER operation: deduct from sender and add to recipient
				if err := c.applyTransferOperation(state, tx.From, op); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown operation type: %s", op.Type)
			}
		}

		// Update nonce
		if state == c.state && tx.From != GenesisAddress {
			c.nonces[tx.From] = tx.Nonce + 1
		}
	}

	return nil
}

// applyMintOperation applies a MINT operation to state
func (c *Chain) applyMintOperation(state *State, op *KVOperation) error {
	// Get current balance
	currentData, _ := state.Get(op.Key)
	currentBalance, err := BalanceFromBytes(currentData)
	if err != nil {
		currentBalance = NewBalance(big.NewInt(0))
	}

	// Add minted amount
	mintAmount := new(big.Int).SetBytes(op.Value)
	currentBalance.Add(mintAmount)

	// Save new balance
	newData := currentBalance.ToBytes()
	state.Set(op.Key, newData)

	// Persist to storage if this is the actual state
	if state == c.state {
		if err := c.storage.SaveState(op.Key, newData); err != nil {
			return fmt.Errorf("failed to save minted balance: %w", err)
		}
	}

	return nil
}

// applyTransferOperation applies a TRANSFER operation to state
// It deducts from sender and adds to recipient
func (c *Chain) applyTransferOperation(state *State, senderAddr string, op *KVOperation) error {
	amount := new(big.Int).SetBytes(op.Value)

	// Deduct from sender
	senderKey := BalanceKey(senderAddr)
	senderData, _ := state.Get(senderKey)
	senderBalance, err := BalanceFromBytes(senderData)
	if err != nil {
		senderBalance = NewBalance(big.NewInt(0))
	}

	if err := senderBalance.Sub(amount); err != nil {
		return fmt.Errorf("insufficient balance for transfer: %w", err)
	}

	state.Set(senderKey, senderBalance.ToBytes())
	if state == c.state {
		if err := c.storage.SaveState(senderKey, senderBalance.ToBytes()); err != nil {
			return fmt.Errorf("failed to save sender balance: %w", err)
		}
	}

	// Add to recipient (op.Key is the recipient's balance key)
	recipientData, _ := state.Get(op.Key)
	recipientBalance, err := BalanceFromBytes(recipientData)
	if err != nil {
		recipientBalance = NewBalance(big.NewInt(0))
	}

	recipientBalance.Add(amount)

	state.Set(op.Key, recipientBalance.ToBytes())
	if state == c.state {
		if err := c.storage.SaveState(op.Key, recipientBalance.ToBytes()); err != nil {
			return fmt.Errorf("failed to save recipient balance: %w", err)
		}
	}

	return nil
}

// ApplyTransactionsWithFees applies transactions with gas fee deduction and collection
// Returns total fees collected and any error
func (c *Chain) ApplyTransactionsWithFees(state *State, transactions []*Transaction, blockProducer string) (*big.Int, error) {
	totalFees := big.NewInt(0)

	for _, tx := range transactions {
		// Skip fee deduction for genesis transactions
		if !tx.IsGenesisTransaction() && c.gasConfig != nil {
			txSize := tx.Size()
			gasFee := c.gasConfig.CalculateGasFee(txSize)

			// Deduct fee from sender
			senderKey := BalanceKey(tx.From)
			senderData, _ := state.Get(senderKey)
			senderBalance, err := BalanceFromBytes(senderData)
			if err != nil {
				senderBalance = NewBalance(big.NewInt(0))
			}

			if err := senderBalance.Sub(gasFee); err != nil {
				return nil, fmt.Errorf("tx %s: insufficient balance for gas: %w", tx.HashString(), err)
			}

			state.Set(senderKey, senderBalance.ToBytes())
			if state == c.state {
				if err := c.storage.SaveState(senderKey, senderBalance.ToBytes()); err != nil {
					return nil, fmt.Errorf("failed to save sender balance: %w", err)
				}
			}

			totalFees.Add(totalFees, gasFee)
		}

		// Apply operations
		for _, op := range tx.Data.Operations {
			// Check authority for MINT operations
			if op.Type == OpTypeMint && !tx.IsGenesisTransaction() {
				if !c.IsAuthority(tx.From) {
					return nil, fmt.Errorf("tx %s: only authorities can mint tokens", tx.HashString())
				}
			}

			switch op.Type {
			case OpTypeSet:
				state.Set(op.Key, op.Value)
				if state == c.state {
					if err := c.storage.SaveState(op.Key, op.Value); err != nil {
						return nil, fmt.Errorf("failed to save state: %w", err)
					}
				}
			case OpTypeDelete:
				state.Delete(op.Key)
				if state == c.state {
					if err := c.storage.DeleteState(op.Key); err != nil {
						return nil, fmt.Errorf("failed to delete state: %w", err)
					}
				}
			case OpTypeMint:
				if err := c.applyMintOperation(state, op); err != nil {
					return nil, err
				}
			case OpTypeTransfer:
				if err := c.applyTransferOperation(state, tx.From, op); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unknown operation type: %s", op.Type)
			}
		}

		// Update nonce
		if state == c.state && !tx.IsGenesisTransaction() {
			c.nonces[tx.From] = tx.Nonce + 1
		}
	}

	// Credit fees to block producer
	if blockProducer != "" && blockProducer != GenesisAddress && totalFees.Sign() > 0 {
		producerKey := BalanceKey(blockProducer)
		producerData, _ := state.Get(producerKey)
		producerBalance, err := BalanceFromBytes(producerData)
		if err != nil {
			producerBalance = NewBalance(big.NewInt(0))
		}
		producerBalance.Add(totalFees)

		state.Set(producerKey, producerBalance.ToBytes())
		if state == c.state {
			if err := c.storage.SaveState(producerKey, producerBalance.ToBytes()); err != nil {
				return nil, fmt.Errorf("failed to save producer balance: %w", err)
			}
		}
	}

	return totalFees, nil
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

// CalculateStateRootWithTransactions calculates what the state root will be
// after applying the given transactions, without modifying the actual state
func (c *Chain) CalculateStateRootWithTransactions(transactions []*Transaction) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Clone current state
	tempState := c.state.Clone()

	// Apply transactions to temporary state
	if err := c.applyTransactionsToState(tempState, transactions); err != nil {
		return nil, err
	}

	// Calculate and return the root
	return tempState.CalculateRoot(), nil
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

// IsAuthority checks if an address is an authority
func (c *Chain) IsAuthority(address string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	normalizedAddr := strings.ToLower(address)
	for _, auth := range c.authorities {
		if strings.ToLower(auth) == normalizedAddr {
			return true
		}
	}
	return false
}

// GetBalance returns the balance for an address
func (c *Chain) GetBalance(address string) (*big.Int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	balanceKey := BalanceKey(address)
	data, exists := c.state.Get(balanceKey)
	if !exists {
		return big.NewInt(0), nil
	}

	balance, err := BalanceFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance.Amount, nil
}

// GetBalanceFromStorage returns the balance for an address from storage
func (c *Chain) GetBalanceFromStorage(address string) (*big.Int, error) {
	balanceKey := BalanceKey(address)
	data, err := c.storage.GetState(balanceKey)
	if err != nil {
		return big.NewInt(0), nil // No balance found
	}

	balance, err := BalanceFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance.Amount, nil
}

// EstimateGasFee estimates the gas fee for a transaction of given size
func (c *Chain) EstimateGasFee(txSize int) *GasEstimate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.gasConfig == nil {
		return &GasEstimate{
			TransactionSize: txSize,
			BaseFee:         big.NewInt(0),
			SizeFee:         big.NewInt(0),
			TotalFee:        big.NewInt(0),
		}
	}

	return c.gasConfig.EstimateGas(txSize)
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
