package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
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

// ValidateTransactionBalance validates that a sender has enough balance for gas fee
func ValidateTransactionBalance(tx *Transaction, senderBalance *big.Int, gasConfig *GasConfig) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	// Genesis transactions don't require balance
	if tx.IsGenesisTransaction() {
		return nil
	}

	// If no gas config, no balance required
	if gasConfig == nil || gasConfig.IsZeroFee() {
		return nil
	}

	// Calculate gas fee
	txSize := tx.Size()
	gasFee := gasConfig.CalculateGasFee(txSize)

	// Check if sender has enough balance
	if senderBalance == nil {
		senderBalance = big.NewInt(0)
	}

	if senderBalance.Cmp(gasFee) < 0 {
		return fmt.Errorf("insufficient balance for gas: have %s, need %s",
			senderBalance.String(), gasFee.String())
	}

	return nil
}

// ValidateMintOperation validates a MINT operation
func ValidateMintOperation(tx *Transaction, authorities []string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	// Genesis transactions can mint
	if tx.IsGenesisTransaction() {
		return nil
	}

	// Check if transaction has MINT operations
	if !tx.HasMintOperations() {
		return nil
	}

	// Check if sender is an authority
	normalizedFrom := strings.ToLower(tx.From)
	isAuth := false
	for _, auth := range authorities {
		if strings.ToLower(auth) == normalizedFrom {
			isAuth = true
			break
		}
	}

	if !isAuth {
		return fmt.Errorf("only authorities can mint tokens, %s is not an authority", tx.From)
	}

	return nil
}

// ValidateTransactionWithChain performs full transaction validation including balance check
func ValidateTransactionWithChain(tx *Transaction, currentNonce uint64, senderBalance *big.Int, gasConfig *GasConfig, authorities []string) error {
	// Basic validation
	if err := ValidateTransaction(tx, currentNonce); err != nil {
		return err
	}

	// Balance validation (gas fees)
	if err := ValidateTransactionBalance(tx, senderBalance, gasConfig); err != nil {
		return err
	}

	// Transfer balance validation
	if err := ValidateTransferBalance(tx, senderBalance, gasConfig); err != nil {
		return err
	}

	// MINT operation validation
	if err := ValidateMintOperation(tx, authorities); err != nil {
		return err
	}

	return nil
}

// ValidateTransferBalance validates that a sender has enough balance for transfers + gas
func ValidateTransferBalance(tx *Transaction, senderBalance *big.Int, gasConfig *GasConfig) error {
	if tx == nil || tx.Data == nil {
		return nil
	}

	// Genesis transactions don't require balance check
	if tx.IsGenesisTransaction() {
		return nil
	}

	// Calculate total transfer amount
	totalTransfer := big.NewInt(0)
	for _, op := range tx.Data.Operations {
		if op.Type == OpTypeTransfer {
			amount := new(big.Int).SetBytes(op.Value)
			totalTransfer.Add(totalTransfer, amount)
		}
	}

	// If no transfers, nothing to validate
	if totalTransfer.Sign() == 0 {
		return nil
	}

	// Calculate gas fee
	gasFee := big.NewInt(0)
	if gasConfig != nil && !gasConfig.IsZeroFee() {
		gasFee = gasConfig.CalculateGasFee(tx.Size())
	}

	// Total required: transfer amount + gas fee
	totalRequired := new(big.Int).Add(totalTransfer, gasFee)

	if senderBalance == nil {
		senderBalance = big.NewInt(0)
	}

	if senderBalance.Cmp(totalRequired) < 0 {
		return fmt.Errorf("insufficient balance for transfer: have %s, need %s (transfer: %s, gas: %s)",
			senderBalance.String(), totalRequired.String(), totalTransfer.String(), gasFee.String())
	}

	return nil
}
