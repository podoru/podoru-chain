# Storage Layer (BadgerDB)

Podoru Chain uses BadgerDB, a high-performance embedded key-value database written in Go, for persistent storage.

## Overview

The storage layer handles:
- Blockchain data persistence
- State management
- Transaction storage
- Block indexing

### Why BadgerDB?

**Advantages**:
- **Pure Go**: No C dependencies, easy deployment
- **High Performance**: LSM tree-based, optimized for SSDs
- **ACID Transactions**: Atomic operations, consistency guarantees
- **Embedded**: No separate database server needed
- **Memory Mapped**: Efficient file I/O
- **Compression**: Built-in compression support

**Comparisons**:
- **vs LevelDB**: Faster reads, better Go integration
- **vs RocksDB**: No CGo, simpler deployment
- **vs BoltDB**: Better write performance, more features

## Storage Architecture

```
Storage Layer
├── Block Store          # Blockchain blocks
│   ├── block:height:N → Block data
│   ├── block:hash:X   → Block data
│   └── meta:height    → Latest height
├── Transaction Store    # Individual transactions
│   └── tx:hash:X      → Transaction data
├── State Store          # Current state (KV pairs)
│   ├── state:key1     → value1
│   ├── state:key2     → value2
│   └── ...
└── Index Store          # Various indexes
    └── height:hash    → Block hash for height
```

## Data Organization

### Key Namespaces

Podoru Chain uses prefixed keys to organize data:

```
Block Storage:
  block:height:<height>    → Block by height
  block:hash:<hash>        → Block by hash

Transaction Storage:
  tx:<txhash>              → Transaction data

State Storage:
  state:<key>              → Application state

Metadata:
  meta:height              → Latest block height
  meta:genesis             → Genesis block hash
```

### Serialization

Data is serialized using JSON before storage:

```go
// Block storage
blockData, _ := json.Marshal(block)
db.Put([]byte("block:height:"+height), blockData)

// State storage
db.Put([]byte("state:"+key), value)
```

## State Management

### In-Memory State Cache

For fast queries, state is cached in memory:

```go
type StateManager struct {
    db          *badger.DB
    stateCache  map[string][]byte  // In-memory cache
    mutex       sync.RWMutex
}
```

**Benefits**:
- Fast reads (no disk I/O)
- Immediate state queries
- Batch updates

**Tradeoffs**:
- Memory usage scales with state size
- Must fit in RAM
- Needs rebuild on restart

### State Operations

#### Read State

```go
func (sm *StateManager) Get(key string) ([]byte, error) {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()

    // Check cache first
    if value, ok := sm.stateCache[key]; ok {
        return value, nil
    }

    // Fall back to disk
    return sm.getFromDisk(key)
}
```

#### Write State

```go
func (sm *StateManager) Set(key string, value []byte) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    // Update cache
    sm.stateCache[key] = value

    // Persist to disk
    return sm.db.Update(func(txn *badger.Txn) error {
        return txn.Set([]byte("state:"+key), value)
    })
}
```

#### Batch Operations

```go
func (sm *StateManager) BatchSet(ops []Operation) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    return sm.db.Update(func(txn *badger.Txn) error {
        for _, op := range ops {
            if op.Type == SET {
                sm.stateCache[op.Key] = op.Value
                txn.Set([]byte("state:"+op.Key), op.Value)
            } else if op.Type == DELETE {
                delete(sm.stateCache, op.Key)
                txn.Delete([]byte("state:"+op.Key))
            }
        }
        return nil
    })
}
```

### State Root Calculation

State root is a Merkle tree hash of all state:

```go
func calculateStateRoot(state map[string][]byte) string {
    // Sort keys for deterministic ordering
    keys := make([]string, 0, len(state))
    for k := range state {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // Build Merkle tree
    leaves := make([][]byte, len(keys))
    for i, k := range keys {
        // Hash key-value pair
        data := fmt.Sprintf("%s:%s", k, state[k])
        hash := sha256.Sum256([]byte(data))
        leaves[i] = hash[:]
    }

    return buildMerkleRoot(leaves)
}
```

## Block Storage

### Storing Blocks

Blocks are stored with multiple indexes:

```go
func (bs *BlockStore) SaveBlock(block *Block) error {
    blockData, _ := json.Marshal(block)

    return bs.db.Update(func(txn *badger.Txn) error {
        // Store by height
        heightKey := fmt.Sprintf("block:height:%d", block.Height)
        txn.Set([]byte(heightKey), blockData)

        // Store by hash
        hashKey := fmt.Sprintf("block:hash:%s", block.Hash)
        txn.Set([]byte(hashKey), blockData)

        // Update latest height
        txn.Set([]byte("meta:height"), []byte(fmt.Sprintf("%d", block.Height)))

        return nil
    })
}
```

### Querying Blocks

```go
// Get by height
func (bs *BlockStore) GetBlockByHeight(height uint64) (*Block, error) {
    var block Block
    err := bs.db.View(func(txn *badger.Txn) error {
        key := fmt.Sprintf("block:height:%d", height)
        item, err := txn.Get([]byte(key))
        if err != nil {
            return err
        }
        return item.Value(func(val []byte) error {
            return json.Unmarshal(val, &block)
        })
    })
    return &block, err
}

// Get by hash
func (bs *BlockStore) GetBlockByHash(hash string) (*Block, error) {
    var block Block
    err := bs.db.View(func(txn *badger.Txn) error {
        key := fmt.Sprintf("block:hash:%s", hash)
        item, err := txn.Get([]byte(key))
        if err != nil {
            return err
        }
        return item.Value(func(val []byte) error {
            return json.Unmarshal(val, &block)
        })
    })
    return &block, err
}
```

## Transaction Storage

### Storing Transactions

```go
func (ts *TransactionStore) SaveTransaction(tx *Transaction) error {
    txData, _ := json.Marshal(tx)

    return ts.db.Update(func(txn *badger.Txn) error {
        key := fmt.Sprintf("tx:%s", tx.Hash)
        return txn.Set([]byte(key), txData)
    })
}
```

### Querying Transactions

```go
func (ts *TransactionStore) GetTransaction(hash string) (*Transaction, error) {
    var tx Transaction
    err := ts.db.View(func(txn *badger.Txn) error {
        key := fmt.Sprintf("tx:%s", hash)
        item, err := txn.Get([]byte(key))
        if err != nil {
            return err
        }
        return item.Value(func(val []byte) error {
            return json.Unmarshal(val, &tx)
        })
    })
    return &tx, err
}
```

## Database Configuration

### BadgerDB Options

```go
func OpenDatabase(dataDir string) (*badger.DB, error) {
    opts := badger.DefaultOptions(dataDir)

    // Performance tuning
    opts.ValueLogFileSize = 256 << 20  // 256 MB
    opts.NumVersionsToKeep = 1         // No versioning needed
    opts.NumLevelZeroTables = 5
    opts.NumLevelZeroTablesStall = 10

    // Memory optimization
    opts.TableBuilderOptions.MaxTableSize = 8 << 20  // 8 MB
    opts.ValueLogMaxEntries = 1000000

    // Compression
    opts.Compression = options.Snappy

    return badger.Open(opts)
}
```

### Directory Structure

```
data/
├── 000000.vlog          # Value log files
├── 000001.sst           # Sorted string tables
├── 000002.sst
├── MANIFEST            # Database manifest
├── KEYREGISTRY         # Key registry
└── ...
```

## Performance Optimization

### Read Optimization

**In-Memory Cache**:
- Hot state data cached in RAM
- Zero disk I/O for cache hits
- Configurable cache size

**Bloom Filters**:
- BadgerDB uses bloom filters
- Fast negative lookups
- Reduces disk reads

**Batch Reads**:
```go
func (sm *StateManager) BatchGet(keys []string) map[string][]byte {
    results := make(map[string][]byte)

    sm.db.View(func(txn *badger.Txn) error {
        for _, key := range keys {
            item, err := txn.Get([]byte("state:" + key))
            if err == nil {
                item.Value(func(val []byte) error {
                    results[key] = val
                    return nil
                })
            }
        }
        return nil
    })

    return results
}
```

### Write Optimization

**Batch Writes**:
```go
func (bs *BlockStore) SaveBlockWithTransactions(block *Block) error {
    return bs.db.Update(func(txn *badger.Txn) error {
        // Save block
        blockData, _ := json.Marshal(block)
        txn.Set([]byte("block:height:"+block.Height), blockData)

        // Save all transactions in same transaction
        for _, tx := range block.Transactions {
            txData, _ := json.Marshal(tx)
            txn.Set([]byte("tx:"+tx.Hash), txData)
        }

        return nil
    })
}
```

**Write-Ahead Log (WAL)**:
- BadgerDB includes built-in WAL
- Crash recovery
- Durability guarantees

### Query Optimization

**Prefix Scans**:
```go
func (sm *StateManager) QueryPrefix(prefix string) map[string][]byte {
    results := make(map[string][]byte)

    sm.db.View(func(txn *badger.Txn) error {
        opts := badger.DefaultIteratorOptions
        opts.Prefix = []byte("state:" + prefix)

        it := txn.NewIterator(opts)
        defer it.Close()

        for it.Seek(opts.Prefix); it.ValidForPrefix(opts.Prefix); it.Next() {
            item := it.Item()
            key := string(item.Key())
            item.Value(func(val []byte) error {
                results[strings.TrimPrefix(key, "state:")] = val
                return nil
            })
        }
        return nil
    })

    return results
}
```

## Garbage Collection

### Value Log GC

BadgerDB requires periodic garbage collection:

```go
func runGarbageCollection(db *badger.DB) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        // Run GC
        err := db.RunValueLogGC(0.5)  // Discard threshold
        if err != nil && err != badger.ErrNoRewrite {
            log.Printf("GC error: %v", err)
        }
    }
}
```

**When to Run GC**:
- Periodically (every 5-10 minutes)
- After large batch operations
- When disk space is low

## Backup and Recovery

### Backup

```go
func BackupDatabase(db *badger.DB, backupPath string) error {
    f, err := os.Create(backupPath)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = db.Backup(f, 0)  // Since timestamp
    return err
}
```

**Usage**:
```bash
# Automated backup
./bin/podoru-node -config config.yaml -backup /backups/backup-$(date +%Y%m%d).bak
```

### Recovery

```go
func RestoreDatabase(db *badger.DB, backupPath string) error {
    f, err := os.Open(backupPath)
    if err != nil {
        return err
    }
    defer f.Close()

    return db.Load(f, 256)  // Max pending writes
}
```

**Usage**:
```bash
# Stop node
systemctl stop podoru-node

# Restore from backup
./bin/podoru-restore -db /data -backup /backups/backup-20240106.bak

# Start node
systemctl start podoru-node
```

## Monitoring

### Database Metrics

```go
// Get database stats
stats := db.Stats()
log.Printf("LSM Size: %d MB", stats.LSMSize/1024/1024)
log.Printf("VLog Size: %d MB", stats.VLogSize/1024/1024)
log.Printf("Pending Writes: %d", stats.PendingWrites)
```

### Health Checks

```bash
# Check database size
du -sh data/

# Check number of files
ls data/*.sst | wc -l

# Monitor I/O
iostat -x 1
```

## Troubleshooting

### Database Corruption

If database becomes corrupted:

```bash
# Stop node
systemctl stop podoru-node

# Try repair
badger repair --dir /path/to/data

# If repair fails, restore from backup
./bin/podoru-restore -db /data -backup /backups/latest.bak

# Start node
systemctl start podoru-node
```

### High Disk Usage

```bash
# Run manual GC
curl -X POST http://localhost:8545/api/v1/admin/gc

# Check for stale data
du -sh data/*

# Compact database
badger compact --dir /path/to/data
```

### Slow Queries

1. Enable query logging
2. Check cache hit rate
3. Optimize state size
4. Add indexes if needed

## Best Practices

### Do's

- Run periodic garbage collection
- Monitor disk usage
- Regular backups
- Use batch operations for writes
- Enable compression

### Don'ts

- Don't run out of disk space
- Don't corrupt database with manual edits
- Don't disable WAL (durability)
- Don't skip backups
- Don't ignore GC errors

## Future Enhancements

### State Pruning

Remove old state to reduce disk usage:
```go
// Future feature
PruneStateBefore(height uint64)
```

### Sharding

Distribute state across multiple databases:
```go
// Future feature
ShardByPrefix(prefix string) *badger.DB
```

### Compression

Improved compression algorithms:
- Zstd for better ratios
- Adaptive compression based on data type

## Further Reading

- [Architecture Overview](README.md)
- [Consensus Mechanism](consensus.md)
- [BadgerDB Documentation](https://dgraph.io/docs/badger/)
- [Performance Tuning](../troubleshooting/README.md)
