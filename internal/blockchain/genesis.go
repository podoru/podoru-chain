package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// GenesisConfig defines the genesis block configuration
type GenesisConfig struct {
	Timestamp       int64             `json:"timestamp"`
	Authorities     []string          `json:"authorities"`
	InitialState    map[string]string `json:"initial_state"`
	TokenConfig     *TokenConfig      `json:"token_config,omitempty"`
	GasConfig       *GasConfigJSON    `json:"gas_config,omitempty"`
	InitialBalances map[string]string `json:"initial_balances,omitempty"` // address -> amount in wei
}

// LoadGenesisConfig loads genesis configuration from a file
func LoadGenesisConfig(filePath string) (*GenesisConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %w", err)
	}

	var config GenesisConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse genesis file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid genesis config: %w", err)
	}

	return &config, nil
}

// Validate validates the genesis configuration
func (gc *GenesisConfig) Validate() error {
	if len(gc.Authorities) == 0 {
		return errors.New("no authorities specified")
	}

	// Check for duplicate authorities
	seen := make(map[string]bool)
	for _, addr := range gc.Authorities {
		if seen[addr] {
			return fmt.Errorf("duplicate authority: %s", addr)
		}
		seen[addr] = true
	}

	// Validate token config if present
	if gc.TokenConfig != nil {
		if err := gc.TokenConfig.Validate(); err != nil {
			return fmt.Errorf("invalid token config: %w", err)
		}
	}

	// Validate gas config if present
	if gc.GasConfig != nil {
		gasConfig, err := GasConfigFromJSON(gc.GasConfig)
		if err != nil {
			return fmt.Errorf("invalid gas config: %w", err)
		}
		if err := gasConfig.Validate(); err != nil {
			return fmt.Errorf("invalid gas config: %w", err)
		}
	}

	// Validate initial balances if present
	if gc.InitialBalances != nil {
		for addr, amountStr := range gc.InitialBalances {
			if _, err := NewBalanceFromString(amountStr); err != nil {
				return fmt.Errorf("invalid balance for %s: %w", addr, err)
			}
		}
	}

	return nil
}

// HasTokenConfig returns true if the genesis has token configuration
func (gc *GenesisConfig) HasTokenConfig() bool {
	return gc.TokenConfig != nil
}

// GetGasConfig returns the gas configuration or default if not set
func (gc *GenesisConfig) GetGasConfig() *GasConfig {
	if gc.GasConfig == nil {
		return nil // No gas fees for legacy genesis
	}
	config, err := GasConfigFromJSON(gc.GasConfig)
	if err != nil {
		return DefaultGasConfig()
	}
	return config
}

// CreateGenesisBlock creates the genesis block from configuration
func CreateGenesisBlock(config *GenesisConfig) *Block {
	// Create initial state transactions
	// Sort keys to ensure deterministic order (maps have random iteration order)
	keys := make([]string, 0, len(config.InitialState))
	for key := range config.InitialState {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var transactions []*Transaction
	var nonce uint64 = 0

	// Create SET transactions for initial state
	for _, key := range keys {
		value := config.InitialState[key]
		tx := &Transaction{
			From:      GenesisAddress,
			Timestamp: config.Timestamp,
			Data: &TransactionData{
				Operations: []*KVOperation{
					{
						Type:  OpTypeSet,
						Key:   key,
						Value: []byte(value),
					},
				},
			},
			Nonce:     nonce,
			Signature: []byte{}, // Genesis transactions are not signed
		}
		tx.ID = tx.Hash()
		transactions = append(transactions, tx)
		nonce++
	}

	// Create MINT transactions for initial balances
	if config.InitialBalances != nil {
		// Sort addresses for deterministic order
		addresses := make([]string, 0, len(config.InitialBalances))
		for addr := range config.InitialBalances {
			addresses = append(addresses, addr)
		}
		sort.Strings(addresses)

		for _, addr := range addresses {
			amountStr := config.InitialBalances[addr]
			balance, err := NewBalanceFromString(amountStr)
			if err != nil {
				continue // Skip invalid balances (already validated)
			}

			tx := &Transaction{
				From:      GenesisAddress,
				Timestamp: config.Timestamp,
				Data: &TransactionData{
					Operations: []*KVOperation{
						{
							Type:  OpTypeMint,
							Key:   BalanceKey(addr),
							Value: balance.ToBytes(),
						},
					},
				},
				Nonce:     nonce,
				Signature: []byte{},
			}
			tx.ID = tx.Hash()
			transactions = append(transactions, tx)
			nonce++
		}
	}

	// Calculate merkle root
	merkleRoot := CalculateMerkleRoot(transactions)

	// Determine block version based on token config
	version := uint32(1)
	if config.HasTokenConfig() {
		version = 2 // Version 2 indicates gas fees enabled
	}

	// Create genesis header
	header := &BlockHeader{
		Version:      version,
		Height:       0,
		PreviousHash: make([]byte, 32), // All zeros for genesis
		Timestamp:    config.Timestamp,
		MerkleRoot:   merkleRoot,
		StateRoot:    make([]byte, 32), // Will be calculated after applying txs
		ProducerAddr: GenesisAddress,
		Nonce:        0,
	}

	// Create genesis block
	block := &Block{
		Header:       header,
		Transactions: transactions,
		Signature:    []byte{}, // Genesis block has no signature
	}

	return block
}

// IsGenesisBlock checks if a block is the genesis block
func IsGenesisBlock(block *Block) bool {
	return block != nil && block.Header.Height == 0
}
