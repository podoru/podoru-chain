package storage

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/podoru/podoru-chain/internal/blockchain"
)

// Key prefixes for different data types
const (
	blockPrefix       = "blk:"       // Block by hash
	blockHeightPrefix = "blh:"       // Block hash by height
	txPrefix          = "tx:"        // Transaction by hash
	statePrefix       = "st:"        // State key-value pairs
	metaPrefix        = "meta:"      // Metadata
	metaHeightKey     = "meta:height" // Current block height
)

// BadgerStore implements blockchain.Storage using BadgerDB
type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore creates a new BadgerDB storage
func NewBadgerStore(dataDir string) (*BadgerStore, error) {
	// Create full path
	dbPath := filepath.Join(dataDir, "badger")

	// Configure BadgerDB options
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable badger's logger for now

	// Open database
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

// SaveBlock saves a block to storage
func (bs *BadgerStore) SaveBlock(block *blockchain.Block) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		// Serialize block
		blockBytes, err := json.Marshal(block)
		if err != nil {
			return fmt.Errorf("failed to marshal block: %w", err)
		}

		// Save by hash
		blockHash := block.Hash()
		hashKey := blockPrefix + hex.EncodeToString(blockHash)
		if err := txn.Set([]byte(hashKey), blockBytes); err != nil {
			return fmt.Errorf("failed to save block by hash: %w", err)
		}

		// Save height -> hash mapping
		heightKey := fmt.Sprintf("%s%020d", blockHeightPrefix, block.Header.Height)
		if err := txn.Set([]byte(heightKey), blockHash); err != nil {
			return fmt.Errorf("failed to save block height index: %w", err)
		}

		return nil
	})
}

// GetBlock retrieves a block by hash
func (bs *BadgerStore) GetBlock(hash []byte) (*blockchain.Block, error) {
	var block blockchain.Block

	err := bs.db.View(func(txn *badger.Txn) error {
		key := blockPrefix + hex.EncodeToString(hash)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &block)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, errors.New("block not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return &block, nil
}

// GetBlockByHeight retrieves a block by height
func (bs *BadgerStore) GetBlockByHeight(height uint64) (*blockchain.Block, error) {
	var blockHash []byte

	// First, get the block hash for this height
	err := bs.db.View(func(txn *badger.Txn) error {
		heightKey := fmt.Sprintf("%s%020d", blockHeightPrefix, height)
		item, err := txn.Get([]byte(heightKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			blockHash = append([]byte{}, val...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("block at height %d not found", height)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	// Now get the block by hash
	return bs.GetBlock(blockHash)
}

// SaveTransaction saves a transaction to storage
func (bs *BadgerStore) SaveTransaction(tx *blockchain.Transaction) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		// Serialize transaction
		txBytes, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction: %w", err)
		}

		// Save by hash
		key := txPrefix + hex.EncodeToString(tx.ID)
		if err := txn.Set([]byte(key), txBytes); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		return nil
	})
}

// GetTransaction retrieves a transaction by hash
func (bs *BadgerStore) GetTransaction(hash []byte) (*blockchain.Transaction, error) {
	var tx blockchain.Transaction

	err := bs.db.View(func(txn *badger.Txn) error {
		key := txPrefix + hex.EncodeToString(hash)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &tx)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, errors.New("transaction not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &tx, nil
}

// SaveState saves a state key-value pair
func (bs *BadgerStore) SaveState(key string, value []byte) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		stateKey := statePrefix + key
		return txn.Set([]byte(stateKey), value)
	})
}

// GetState retrieves a state value by key
func (bs *BadgerStore) GetState(key string) ([]byte, error) {
	var value []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		stateKey := statePrefix + key
		item, err := txn.Get([]byte(stateKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, errors.New("state key not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	return value, nil
}

// DeleteState deletes a state key
func (bs *BadgerStore) DeleteState(key string) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		stateKey := statePrefix + key
		return txn.Delete([]byte(stateKey))
	})
}

// SaveBlockHeight saves the current block height
func (bs *BadgerStore) SaveBlockHeight(height uint64) error {
	return bs.db.Update(func(txn *badger.Txn) error {
		heightBytes := []byte(fmt.Sprintf("%d", height))
		return txn.Set([]byte(metaHeightKey), heightBytes)
	})
}

// GetLatestBlockHeight retrieves the latest block height
func (bs *BadgerStore) GetLatestBlockHeight() (uint64, error) {
	var height uint64

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(metaHeightKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			_, err := fmt.Sscanf(string(val), "%d", &height)
			return err
		})
	})

	if err == badger.ErrKeyNotFound {
		return 0, errors.New("height not found")
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get height: %w", err)
	}

	return height, nil
}

// Close closes the database
func (bs *BadgerStore) Close() error {
	return bs.db.Close()
}

// RunGC runs garbage collection on the database
func (bs *BadgerStore) RunGC(discardRatio float64) error {
	return bs.db.RunValueLogGC(discardRatio)
}

// ScanStateByPrefix scans all state keys with a given prefix
func (bs *BadgerStore) ScanStateByPrefix(prefix string, limit int) (map[string][]byte, error) {
	results := make(map[string][]byte)
	count := 0

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(statePrefix + prefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			if limit > 0 && count >= limit {
				break
			}

			item := it.Item()
			key := string(item.Key())

			// Remove the statePrefix to get the actual key
			actualKey := key[len(statePrefix):]

			err := item.Value(func(val []byte) error {
				results[actualKey] = append([]byte{}, val...)
				return nil
			})

			if err != nil {
				return err
			}

			count++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan by prefix: %w", err)
	}

	return results, nil
}

// GetAllStateKeys returns all state keys (useful for debugging, use carefully)
func (bs *BadgerStore) GetAllStateKeys(limit int) ([]string, error) {
	var keys []string
	count := 0

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(statePrefix)
		opts.PrefetchValues = false // We only need keys

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			if limit > 0 && count >= limit {
				break
			}

			item := it.Item()
			key := string(item.Key())

			// Remove the statePrefix to get the actual key
			actualKey := key[len(statePrefix):]
			keys = append(keys, actualKey)
			count++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get state keys: %w", err)
	}

	return keys, nil
}
