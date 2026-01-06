package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/podoru/podoru-chain/internal/crypto"
)

// BlockHeader contains block metadata
type BlockHeader struct {
	Version      uint32 `json:"version"`
	Height       uint64 `json:"height"`
	PreviousHash []byte `json:"previous_hash"`
	Timestamp    int64  `json:"timestamp"`      // Unix timestamp
	MerkleRoot   []byte `json:"merkle_root"`    // Root of tx merkle tree
	StateRoot    []byte `json:"state_root"`     // Root hash of KV state
	ProducerAddr string `json:"producer_addr"`  // Block producer address
	Nonce        uint64 `json:"nonce"`          // Can be used for ordering
}

// Block represents a single block in the blockchain
type Block struct {
	Header       *BlockHeader   `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	Signature    []byte         `json:"signature"` // PoA signature
}

// NewBlock creates a new block
func NewBlock(header *BlockHeader, transactions []*Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: transactions,
	}
}

// Hash calculates the block hash (hash of the header)
func (b *Block) Hash() []byte {
	headerBytes, err := json.Marshal(b.Header)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal block header: %v", err))
	}

	hash := sha256.Sum256(headerBytes)
	return hash[:]
}

// Sign signs the block with a private key
func (b *Block) Sign(privateKey *ecdsa.PrivateKey) error {
	hash := b.Hash()

	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}

	b.Signature = signature
	return nil
}

// Verify verifies the block signature
func (b *Block) Verify() error {
	if b.Signature == nil || len(b.Signature) == 0 {
		// Genesis block has no signature
		if b.Header.Height == 0 {
			return nil
		}
		return errors.New("block has no signature")
	}

	hash := b.Hash()

	// Recover address from signature
	recoveredAddr, err := crypto.RecoverAddress(hash, b.Signature)
	if err != nil {
		return fmt.Errorf("failed to recover address: %w", err)
	}

	// Normalize addresses for comparison
	normalizedProducer := crypto.NormalizeAddress(b.Header.ProducerAddr)
	normalizedRecovered := crypto.NormalizeAddress(recoveredAddr)

	if normalizedProducer != normalizedRecovered {
		return fmt.Errorf("signature verification failed: expected %s, got %s",
			normalizedProducer, normalizedRecovered)
	}

	return nil
}

// CalculateMerkleRoot calculates the merkle root of transactions
func CalculateMerkleRoot(transactions []*Transaction) []byte {
	if len(transactions) == 0 {
		return make([]byte, 32) // Empty hash
	}

	// Get transaction hashes
	hashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Hash()
	}

	// Build merkle tree bottom-up
	return buildMerkleTree(hashes)
}

// buildMerkleTree builds a merkle tree from a list of hashes
func buildMerkleTree(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return make([]byte, 32)
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	var nextLevel [][]byte
	for i := 0; i < len(hashes); i += 2 {
		if i+1 < len(hashes) {
			// Hash pair together
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash[:])
		} else {
			// Odd number, hash with itself
			combined := append(hashes[i], hashes[i]...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash[:])
		}
	}

	return buildMerkleTree(nextLevel)
}

// Size returns the approximate size of the block in bytes
func (b *Block) Size() int {
	blockBytes, err := json.Marshal(b)
	if err != nil {
		return 0
	}
	return len(blockBytes)
}

// TransactionCount returns the number of transactions in the block
func (b *Block) TransactionCount() int {
	return len(b.Transactions)
}

// GetTransactionByHash finds a transaction in the block by its hash
func (b *Block) GetTransactionByHash(hash []byte) *Transaction {
	for _, tx := range b.Transactions {
		if string(tx.ID) == string(hash) {
			return tx
		}
	}
	return nil
}

// HashString returns the block hash as a hex string with 0x prefix
func (b *Block) HashString() string {
	return fmt.Sprintf("0x%x", b.Hash())
}

// PreviousHashString returns the previous block hash as a hex string with 0x prefix
func (b *BlockHeader) PreviousHashString() string {
	return fmt.Sprintf("0x%x", b.PreviousHash)
}
