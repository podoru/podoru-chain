package network

import (
	"errors"
	"sync"

	"github.com/podoru/podoru-chain/internal/blockchain"
)

const (
	// MaxMempoolSize is the maximum number of transactions in the mempool
	MaxMempoolSize = 10000

	// MaxMempoolTxSize is the maximum size of a single transaction
	MaxMempoolTxSize = 1024 * 1024 // 1 MB
)

// Mempool manages pending transactions
type Mempool struct {
	mu           sync.RWMutex
	transactions map[string]*blockchain.Transaction // txID -> transaction
	byNonce      map[string]map[uint64]*blockchain.Transaction // address -> nonce -> tx
}

// NewMempool creates a new mempool
func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[string]*blockchain.Transaction),
		byNonce:      make(map[string]map[uint64]*blockchain.Transaction),
	}
}

// AddTransaction adds a transaction to the mempool
func (mp *Mempool) AddTransaction(tx *blockchain.Transaction) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Check mempool size
	if len(mp.transactions) >= MaxMempoolSize {
		return errors.New("mempool is full")
	}

	// Check transaction size
	if tx.Size() > MaxMempoolTxSize {
		return errors.New("transaction too large")
	}

	// Check if transaction already exists
	txID := string(tx.ID)
	if _, exists := mp.transactions[txID]; exists {
		return errors.New("transaction already in mempool")
	}

	// Add transaction
	mp.transactions[txID] = tx

	// Index by nonce
	if mp.byNonce[tx.From] == nil {
		mp.byNonce[tx.From] = make(map[uint64]*blockchain.Transaction)
	}
	mp.byNonce[tx.From][tx.Nonce] = tx

	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (mp *Mempool) RemoveTransaction(txID []byte) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	txIDStr := string(txID)
	tx, exists := mp.transactions[txIDStr]
	if !exists {
		return
	}

	delete(mp.transactions, txIDStr)

	if mp.byNonce[tx.From] != nil {
		delete(mp.byNonce[tx.From], tx.Nonce)
		if len(mp.byNonce[tx.From]) == 0 {
			delete(mp.byNonce, tx.From)
		}
	}
}

// RemoveTransactions removes multiple transactions
func (mp *Mempool) RemoveTransactions(transactions []*blockchain.Transaction) {
	for _, tx := range transactions {
		mp.RemoveTransaction(tx.ID)
	}
}

// GetTransaction retrieves a transaction by ID
func (mp *Mempool) GetTransaction(txID []byte) (*blockchain.Transaction, error) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	tx, exists := mp.transactions[string(txID)]
	if !exists {
		return nil, errors.New("transaction not found in mempool")
	}

	return tx, nil
}

// GetPendingTransactions returns pending transactions up to maxCount
func (mp *Mempool) GetPendingTransactions(maxCount int) []*blockchain.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	transactions := make([]*blockchain.Transaction, 0, maxCount)

	for _, tx := range mp.transactions {
		if len(transactions) >= maxCount {
			break
		}
		transactions = append(transactions, tx)
	}

	return transactions
}

// GetAllPendingTransactions returns all pending transactions
func (mp *Mempool) GetAllPendingTransactions() []*blockchain.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	transactions := make([]*blockchain.Transaction, 0, len(mp.transactions))

	for _, tx := range mp.transactions {
		transactions = append(transactions, tx)
	}

	return transactions
}

// Count returns the number of transactions in the mempool
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.transactions)
}

// Clear removes all transactions from the mempool
func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions = make(map[string]*blockchain.Transaction)
	mp.byNonce = make(map[string]map[uint64]*blockchain.Transaction)
}

// HasTransaction checks if a transaction exists in the mempool
func (mp *Mempool) HasTransaction(txID []byte) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	_, exists := mp.transactions[string(txID)]
	return exists
}

// GetTransactionsByAddress returns all transactions from a specific address
func (mp *Mempool) GetTransactionsByAddress(address string) []*blockchain.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txMap, exists := mp.byNonce[address]
	if !exists {
		return []*blockchain.Transaction{}
	}

	transactions := make([]*blockchain.Transaction, 0, len(txMap))
	for _, tx := range txMap {
		transactions = append(transactions, tx)
	}

	return transactions
}
