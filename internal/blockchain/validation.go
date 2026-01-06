package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"time"
)

const (
	// MaxBlockSize is the maximum size of a block in bytes (1 MB)
	MaxBlockSize = 1024 * 1024

	// MaxTransactionsPerBlock is the maximum number of transactions per block
	MaxTransactionsPerBlock = 1000

	// MaxFutureBlockTime is the maximum time a block can be in the future
	MaxFutureBlockTime = 30 // seconds
)

// ValidateBlock performs comprehensive block validation
func ValidateBlock(block *Block, previousBlock *Block, authorities []string) error {
	if block == nil {
		return errors.New("block is nil")
	}

	if block.Header == nil {
		return errors.New("block header is nil")
	}

	// Genesis block validation
	if IsGenesisBlock(block) {
		return validateGenesisBlock(block)
	}

	// Check block size
	if block.Size() > MaxBlockSize {
		return fmt.Errorf("block too large: %d bytes (max %d)", block.Size(), MaxBlockSize)
	}

	// Check transaction count
	if len(block.Transactions) > MaxTransactionsPerBlock {
		return fmt.Errorf("too many transactions: %d (max %d)",
			len(block.Transactions), MaxTransactionsPerBlock)
	}

	// Validate block height
	if previousBlock != nil {
		if block.Header.Height != previousBlock.Header.Height+1 {
			return fmt.Errorf("invalid block height: expected %d, got %d",
				previousBlock.Header.Height+1, block.Header.Height)
		}
	}

	// Validate previous hash
	if previousBlock != nil {
		if !bytes.Equal(block.Header.PreviousHash, previousBlock.Hash()) {
			return errors.New("invalid previous hash")
		}
	}

	// Validate timestamp
	if block.Header.Timestamp > time.Now().Unix()+MaxFutureBlockTime {
		return errors.New("block timestamp too far in future")
	}

	if previousBlock != nil && block.Header.Timestamp <= previousBlock.Header.Timestamp {
		return errors.New("block timestamp must be greater than previous block")
	}

	// Validate block producer is an authority
	isAuthority := false
	for _, addr := range authorities {
		if addr == block.Header.ProducerAddr {
			isAuthority = true
			break
		}
	}
	if !isAuthority {
		return fmt.Errorf("block producer %s is not an authority", block.Header.ProducerAddr)
	}

	// Verify block signature
	if err := block.Verify(); err != nil {
		return fmt.Errorf("block signature verification failed: %w", err)
	}

	// Validate all transactions
	for i, tx := range block.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction at index %d: %w", i, err)
		}
	}

	// Verify merkle root
	calculatedMerkle := CalculateMerkleRoot(block.Transactions)
	if !bytes.Equal(calculatedMerkle, block.Header.MerkleRoot) {
		return errors.New("invalid merkle root")
	}

	return nil
}

// validateGenesisBlock validates the genesis block
func validateGenesisBlock(block *Block) error {
	if block.Header.Height != 0 {
		return errors.New("genesis block must have height 0")
	}

	// Genesis block previous hash should be all zeros
	emptyHash := make([]byte, 32)
	if !bytes.Equal(block.Header.PreviousHash, emptyHash) {
		return errors.New("genesis block must have empty previous hash")
	}

	// Genesis block doesn't require signature
	// Merkle root should still be valid
	calculatedMerkle := CalculateMerkleRoot(block.Transactions)
	if !bytes.Equal(calculatedMerkle, block.Header.MerkleRoot) {
		return errors.New("invalid merkle root in genesis block")
	}

	return nil
}

// ValidateTransaction validates a transaction (called by Transaction.Validate())
// This is a placeholder for any chain-level transaction validation
func ValidateTransaction(tx *Transaction, currentNonce uint64) error {
	if err := tx.Validate(); err != nil {
		return err
	}

	// Check nonce
	if tx.Nonce != currentNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", currentNonce, tx.Nonce)
	}

	return nil
}
