package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/podoru/podoru-chain/internal/crypto"
)

// OperationType defines the type of key-value operation
type OperationType string

const (
	OpTypeSet    OperationType = "SET"
	OpTypeDelete OperationType = "DELETE"
)

// KVOperation represents a single key-value operation
type KVOperation struct {
	Type  OperationType `json:"type"`
	Key   string        `json:"key"`
	Value []byte        `json:"value,omitempty"` // Empty for DELETE
}

// TransactionData contains the actual key-value pairs
type TransactionData struct {
	Operations []*KVOperation `json:"operations"`
}

// Transaction represents a key-value operation on the blockchain
type Transaction struct {
	ID        []byte           `json:"id"`         // Transaction hash
	From      string           `json:"from"`       // Sender address
	Timestamp int64            `json:"timestamp"`  // Unix timestamp
	Data      *TransactionData `json:"data"`       // Transaction data
	Signature []byte           `json:"signature"`  // Signature
	Nonce     uint64           `json:"nonce"`      // For ordering/replay protection
}

// NewTransaction creates a new transaction
func NewTransaction(from string, timestamp int64, data *TransactionData, nonce uint64) *Transaction {
	tx := &Transaction{
		From:      from,
		Timestamp: timestamp,
		Data:      data,
		Nonce:     nonce,
	}
	tx.ID = tx.Hash()
	return tx
}

// Hash calculates the transaction hash
func (tx *Transaction) Hash() []byte {
	// Create a copy without ID and Signature for hashing
	hashTx := struct {
		From      string           `json:"from"`
		Timestamp int64            `json:"timestamp"`
		Data      *TransactionData `json:"data"`
		Nonce     uint64           `json:"nonce"`
	}{
		From:      tx.From,
		Timestamp: tx.Timestamp,
		Data:      tx.Data,
		Nonce:     tx.Nonce,
	}

	txBytes, err := json.Marshal(hashTx)
	if err != nil {
		// This should never happen with valid data
		panic(fmt.Sprintf("failed to marshal transaction: %v", err))
	}

	hash := sha256.Sum256(txBytes)
	return hash[:]
}

// Sign signs the transaction with a private key
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	hash := tx.Hash()

	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	tx.Signature = signature
	tx.ID = hash
	return nil
}

// Verify verifies the transaction signature
func (tx *Transaction) Verify() error {
	if tx.Signature == nil || len(tx.Signature) == 0 {
		return errors.New("transaction has no signature")
	}

	if tx.ID == nil || len(tx.ID) == 0 {
		return errors.New("transaction has no ID")
	}

	hash := tx.Hash()

	// Recover address from signature
	recoveredAddr, err := crypto.RecoverAddress(hash, tx.Signature)
	if err != nil {
		return fmt.Errorf("failed to recover address: %w", err)
	}

	// Normalize addresses for comparison
	normalizedFrom := crypto.NormalizeAddress(tx.From)
	normalizedRecovered := crypto.NormalizeAddress(recoveredAddr)

	if normalizedFrom != normalizedRecovered {
		return fmt.Errorf("signature verification failed: expected %s, got %s",
			normalizedFrom, normalizedRecovered)
	}

	return nil
}

// Validate performs basic validation on the transaction
func (tx *Transaction) Validate() error {
	// Check required fields
	if tx.From == "" {
		return errors.New("transaction has no sender")
	}

	if !crypto.IsValidAddress(tx.From) {
		return fmt.Errorf("invalid sender address: %s", tx.From)
	}

	if tx.Data == nil {
		return errors.New("transaction has no data")
	}

	if len(tx.Data.Operations) == 0 {
		return errors.New("transaction has no operations")
	}

	// Validate operations
	for i, op := range tx.Data.Operations {
		if op.Key == "" {
			return fmt.Errorf("operation %d has empty key", i)
		}

		if op.Type != OpTypeSet && op.Type != OpTypeDelete {
			return fmt.Errorf("operation %d has invalid type: %s", i, op.Type)
		}

		if op.Type == OpTypeSet && len(op.Value) == 0 {
			return fmt.Errorf("operation %d is SET but has no value", i)
		}

		// Check key and value sizes (prevent DOS)
		const maxKeySize = 1024         // 1 KB
		const maxValueSize = 1024 * 1024 // 1 MB

		if len(op.Key) > maxKeySize {
			return fmt.Errorf("operation %d key too large: %d bytes (max %d)",
				i, len(op.Key), maxKeySize)
		}

		if len(op.Value) > maxValueSize {
			return fmt.Errorf("operation %d value too large: %d bytes (max %d)",
				i, len(op.Value), maxValueSize)
		}
	}

	// Verify signature
	if err := tx.Verify(); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// Size returns the approximate size of the transaction in bytes
func (tx *Transaction) Size() int {
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return 0
	}
	return len(txBytes)
}
