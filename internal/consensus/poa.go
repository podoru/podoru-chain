package consensus

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/podoru/podoru-chain/internal/blockchain"
)

// PoAEngine implements Proof of Authority consensus
type PoAEngine struct {
	mu           sync.RWMutex
	authorities  []string          // List of authority addresses
	authorityMap map[string]bool   // Quick lookup for authorities
	blockTime    time.Duration     // Target block time
}

// NewPoAEngine creates a new PoA consensus engine
func NewPoAEngine(authorities []string, blockTime time.Duration) (*PoAEngine, error) {
	if len(authorities) == 0 {
		return nil, errors.New("no authorities provided")
	}

	if blockTime <= 0 {
		blockTime = 5 * time.Second // Default 5 seconds
	}

	authMap := make(map[string]bool)
	for _, addr := range authorities {
		if authMap[addr] {
			return nil, fmt.Errorf("duplicate authority: %s", addr)
		}
		authMap[addr] = true
	}

	return &PoAEngine{
		authorities:  authorities,
		authorityMap: authMap,
		blockTime:    blockTime,
	}, nil
}

// IsAuthorized checks if an address is an authority
func (poa *PoAEngine) IsAuthorized(address string) bool {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	return poa.authorityMap[address]
}

// GetBlockProducer determines which authority should produce the next block
// Uses simple round-robin based on block height
func (poa *PoAEngine) GetBlockProducer(height uint64) string {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	if len(poa.authorities) == 0 {
		return ""
	}

	index := height % uint64(len(poa.authorities))
	return poa.authorities[index]
}

// CanProduceBlock checks if a given address can produce a block at this height
func (poa *PoAEngine) CanProduceBlock(height uint64, address string) bool {
	expectedProducer := poa.GetBlockProducer(height)
	return expectedProducer == address
}

// ValidateBlockProducer validates that the correct authority produced the block
func (poa *PoAEngine) ValidateBlockProducer(block *blockchain.Block) error {
	// Skip validation for genesis block
	if blockchain.IsGenesisBlock(block) {
		return nil
	}

	poa.mu.RLock()
	defer poa.mu.RUnlock()

	// Check if producer is an authority
	if !poa.authorityMap[block.Header.ProducerAddr] {
		return fmt.Errorf("producer %s is not an authority", block.Header.ProducerAddr)
	}

	// Check if it's the correct producer for this height
	expectedProducer := poa.GetBlockProducer(block.Header.Height)
	if block.Header.ProducerAddr != expectedProducer {
		return fmt.Errorf("wrong producer for height %d: expected %s, got %s",
			block.Header.Height, expectedProducer, block.Header.ProducerAddr)
	}

	return nil
}

// GetBlockTime returns the target block time
func (poa *PoAEngine) GetBlockTime() time.Duration {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	return poa.blockTime
}

// GetAuthorities returns the list of authorities
func (poa *PoAEngine) GetAuthorities() []string {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	// Return a copy to prevent modification
	authorities := make([]string, len(poa.authorities))
	copy(authorities, poa.authorities)
	return authorities
}

// GetAuthorityCount returns the number of authorities
func (poa *PoAEngine) GetAuthorityCount() int {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	return len(poa.authorities)
}

// UpdateAuthorities updates the list of authorities
// Note: In production, this should be done through a governance mechanism
func (poa *PoAEngine) UpdateAuthorities(newAuthorities []string) error {
	if len(newAuthorities) == 0 {
		return errors.New("cannot set empty authority list")
	}

	poa.mu.Lock()
	defer poa.mu.Unlock()

	// Check for duplicates
	authMap := make(map[string]bool)
	for _, addr := range newAuthorities {
		if authMap[addr] {
			return fmt.Errorf("duplicate authority in new list: %s", addr)
		}
		authMap[addr] = true
	}

	poa.authorities = newAuthorities
	poa.authorityMap = authMap

	return nil
}

// CalculateNextBlockTime calculates when the next block should be produced
func (poa *PoAEngine) CalculateNextBlockTime(lastBlockTime int64) time.Time {
	poa.mu.RLock()
	defer poa.mu.RUnlock()

	lastTime := time.Unix(lastBlockTime, 0)
	return lastTime.Add(poa.blockTime)
}

// ShouldProduceBlock checks if it's time to produce a new block
func (poa *PoAEngine) ShouldProduceBlock(lastBlockTime int64) bool {
	nextBlockTime := poa.CalculateNextBlockTime(lastBlockTime)
	return time.Now().After(nextBlockTime)
}
