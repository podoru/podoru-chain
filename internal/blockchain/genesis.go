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
	Timestamp    int64             `json:"timestamp"`
	Authorities  []string          `json:"authorities"`
	InitialState map[string]string `json:"initial_state"`
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

	return nil
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
	for _, key := range keys {
		value := config.InitialState[key]
		tx := &Transaction{
			From:      "0x0000000000000000000000000000000000000000", // Genesis address
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
			Nonce:     0,
			Signature: []byte{}, // Genesis transactions are not signed
		}
		tx.ID = tx.Hash()
		transactions = append(transactions, tx)
	}

	// Calculate merkle root
	merkleRoot := CalculateMerkleRoot(transactions)

	// Create genesis header
	header := &BlockHeader{
		Version:      1,
		Height:       0,
		PreviousHash: make([]byte, 32), // All zeros for genesis
		Timestamp:    config.Timestamp,
		MerkleRoot:   merkleRoot,
		StateRoot:    make([]byte, 32), // Will be calculated after applying txs
		ProducerAddr: "0x0000000000000000000000000000000000000000",
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
