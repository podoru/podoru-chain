# Genesis File Persistence and Node Stability

## Question: Is the node still stable if genesis.json is deleted?

**Short Answer:** ✅ **Yes**, the node remains completely stable after genesis.json is deleted.

## How Genesis Works

### Initial Node Startup (First Time)

When a node starts for the first time:

1. **Checks storage** - Tries to load blockchain from BadgerDB
2. **No blockchain found** - Storage is empty
3. **Loads genesis.json** - Reads the genesis configuration file
4. **Creates genesis block** - Builds the first block with initial state
5. **Saves to BadgerDB** - Persists genesis block in database
6. **Deletes from memory** - Genesis config no longer needed

### Subsequent Node Restarts

When a node restarts after initialization:

1. **Checks storage** - Tries to load blockchain from BadgerDB
2. **Blockchain found** - Loads from database ✅
3. **Genesis.json NOT read** - File is never accessed again
4. **Continues operation** - Works normally

## Code Flow

From `internal/node/node.go`:

```go
func (n *Node) initializeChain() error {
    // Try to load existing chain
    if err := n.chain.LoadFromStorage(); err != nil {
        // Chain doesn't exist, create genesis
        n.logger.Info("Creating genesis block...")

        genesisConfig, err := blockchain.LoadGenesisConfig(n.config.GenesisPath)
        if err != nil {
            return fmt.Errorf("failed to load genesis config: %w", err)
        }

        genesisBlock := blockchain.CreateGenesisBlock(genesisConfig)

        if err := n.chain.Initialize(genesisBlock); err != nil {
            return fmt.Errorf("failed to initialize chain with genesis: %w", err)
        }

        n.logger.Info("Genesis block created")
    } else {
        // Genesis.json is NOT accessed here!
        n.logger.Infof("Loaded blockchain from storage (height: %d)", n.chain.GetHeight())
    }

    return nil
}
```

**Key Point:** The `LoadGenesisConfig()` call only happens inside the error branch - when the blockchain doesn't exist in storage yet.

## Where is Blockchain Data Stored?

All blockchain data is persisted in **BadgerDB**, located at:

```
<data_dir>/badger/
```

This directory contains:
- **Genesis block** - The first block with initial state
- **All subsequent blocks** - Every block added to the chain
- **State data** - All key-value pairs
- **Transactions** - Every transaction
- **Block height** - Current chain height

## Testing This Behavior

### Test 1: Delete genesis.json After Startup

```bash
# Start the network
make docker-compose-up

# Wait for initialization
sleep 10

# Check that blockchain is running
curl http://localhost:8545/api/v1/chain/info

# Stop the node
docker-compose -f docker/docker-compose.yml down

# Delete genesis.json
sudo rm docker/data/producer1/genesis.json

# Restart the node
docker-compose -f docker/docker-compose.yml up -d

# Check logs - should say "Loaded blockchain from storage"
docker-compose -f docker/docker-compose.yml logs producer1 | grep -i "loaded"

# Verify it still works
curl http://localhost:8545/api/v1/chain/info
```

**Expected Result:** ✅ Node starts successfully and continues from where it left off.

### Test 2: Fresh Start Requires genesis.json

```bash
# Clean everything (including BadgerDB)
sudo rm -rf docker/data/producer1/badger/
sudo rm docker/data/producer1/genesis.json

# Try to start
docker-compose -f docker/docker-compose.yml up -d

# Check logs
docker-compose -f docker/docker-compose.yml logs producer1
```

**Expected Result:** ❌ Node fails to start with "failed to load genesis config" error.

## When is genesis.json Required?

Genesis.json is **ONLY** required when:

1. ✅ **First-time initialization** - No blockchain data exists
2. ✅ **Complete reset** - BadgerDB directory is deleted
3. ✅ **New nodes joining** - Need to create identical genesis
4. ❌ **Normal restarts** - NOT required (uses BadgerDB)
5. ❌ **Updates/upgrades** - NOT required (uses BadgerDB)
6. ❌ **Crashes/recovery** - NOT required (uses BadgerDB)

## Best Practices

### For Production Deployments

1. **Keep genesis.json as backup**
   - Store in version control
   - Document the exact configuration
   - Useful for setting up new nodes

2. **Protect BadgerDB directory**
   - This is your actual blockchain data
   - Regular backups of `/data/badger/`
   - Never delete unless intentional reset

3. **Verify genesis hash**
   - All nodes in network must have identical genesis
   - Use SHA256 checksum to verify
   - Document genesis hash in network info

### For Development

1. **Keep genesis.json available**
   - Useful for frequent resets
   - Easy to recreate network
   - Can modify for testing

2. **Use clean-wizard for resets**
   ```bash
   make clean-wizard  # Removes BadgerDB and genesis
   make setup-wizard  # Creates new network
   ```

## Recovery Scenarios

### Scenario 1: Lost genesis.json, BadgerDB intact

**Status:** ✅ **SAFE** - Node continues normally

**Action:** None required. Optionally recreate genesis.json from network info for documentation.

### Scenario 2: Lost BadgerDB, genesis.json intact

**Status:** ⚠️ **BLOCKCHAIN RESET** - All blocks and state lost

**Action:**
1. Node will recreate genesis block from genesis.json
2. Will sync from network peers
3. Full node: Syncs entire blockchain
4. Producer: Cannot produce until synced

### Scenario 3: Lost both genesis.json and BadgerDB

**Status:** ❌ **CANNOT START**

**Action:**
1. Get genesis.json from network operator
2. Verify SHA256 hash matches network
3. Restart node to initialize

### Scenario 4: Corrupted BadgerDB

**Status:** ⚠️ **NEEDS RESET**

**Action:**
```bash
# Stop node
docker-compose down

# Backup current data (optional)
mv badger badger.corrupted

# Delete BadgerDB
rm -rf badger/

# Restart - will recreate from genesis.json
docker-compose up -d
```

## Architecture Decision

**Why is genesis.json separate from BadgerDB?**

1. **Human-readable configuration** - Easy to audit and verify
2. **Network coordination** - Share file to ensure identical genesis
3. **Flexibility** - Can recreate blockchain from scratch
4. **Simplicity** - Clear separation of config vs. data

**Why not embed genesis in code?**

1. **Different networks** - Same binary, different chains
2. **Governance** - Easy to audit what network you're joining
3. **Testing** - Can create test networks with different genesis
4. **Security** - Verify genesis hash before joining network

## Summary

| Scenario | genesis.json Required? | Blockchain Continues? |
|----------|----------------------|---------------------|
| Normal restart | ❌ No | ✅ Yes |
| Crash recovery | ❌ No | ✅ Yes |
| After deletion of genesis.json | ❌ No | ✅ Yes |
| After deletion of BadgerDB | ✅ Yes | ❌ No (resyncs) |
| Fresh installation | ✅ Yes | N/A (new chain) |
| Software upgrade | ❌ No | ✅ Yes |

**Key Takeaway:** Once initialized, genesis.json is only needed as documentation/backup. The actual blockchain state lives in BadgerDB.

## File Importance Ranking

1. **Most Critical:** `badger/` - Your actual blockchain data
2. **Important:** `genesis.json` - Needed for fresh starts and new nodes
3. **Important:** `config.yaml` - Node configuration
4. **Important:** `keys/*.key` - Producer private keys (if producer)

**Backup Strategy:**
```bash
# Essential backup (minimal)
tar -czf blockchain-backup.tar.gz badger/ genesis.json

# Full backup (recommended)
tar -czf blockchain-full-backup.tar.gz badger/ genesis.json config.yaml keys/
```

## Related Documentation

- [Joining a Network](../joining-network.md) - Requires genesis.json for new nodes
- [Troubleshooting](../troubleshooting/README.md) - Recovery procedures
- [Configuration](../configuration/README.md) - Understanding config files
