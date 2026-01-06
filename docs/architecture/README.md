# Architecture Overview

Podoru Chain is built with a modular architecture designed for simplicity, performance, and decentralization.

## Core Components

### 1. Consensus Layer
- **Proof of Authority (PoA)** consensus mechanism
- Round-robin block producer selection
- Deterministic finality
- Configurable block times

See [Consensus Details](consensus.md)

### 2. Storage Layer
- **BadgerDB** embedded key-value database
- In-memory state cache for fast queries
- Merkle tree for state verification
- Efficient batch operations

See [Storage Details](storage.md)

### 3. Networking Layer
- **TCP-based P2P** communication
- Gossip protocol for block propagation
- Peer discovery and management
- Block synchronization

See [Networking Details](networking.md)

### 4. API Layer
- **RESTful HTTP** interface
- JSON request/response format
- Comprehensive endpoint coverage
- Query optimization

See [API Reference](../api-reference/README.md)

## System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    REST API Server                       │
│            (Port 8545 - HTTP/JSON)                       │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────────┐
│                   Node Orchestrator                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Blockchain  │  │  Consensus   │  │    Network   │  │
│  │    Engine    │  │    Engine    │  │    Manager   │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                  │                  │          │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐  │
│  │    State     │  │   Mempool    │  │   P2P Layer  │  │
│  │   Manager    │  │              │  │  (Port 9000) │  │
│  └──────┬───────┘  └──────────────┘  └──────────────┘  │
└─────────┼──────────────────────────────────────────────┘
          │
┌─────────┴────────────────────────────────────────────────┐
│                    Storage Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   BadgerDB   │  │  Block Store │  │  State Store │   │
│  │  (Key-Value) │  │              │  │              │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
└──────────────────────────────────────────────────────────┘
```

## Data Flow

### Block Production Flow (Producer Nodes)

1. **Transaction Collection**
   - Receive transactions via API
   - Validate transaction signatures
   - Add to mempool

2. **Block Creation** (when it's producer's turn)
   - Collect pending transactions from mempool
   - Execute transaction operations
   - Update state and calculate state root
   - Create block with transaction and state Merkle roots

3. **Block Signing**
   - Sign block with producer's private key
   - Add signature to block header

4. **Block Broadcast**
   - Propagate block to all peers via gossip protocol
   - Clear mempool of included transactions

### Block Validation Flow (All Nodes)

1. **Block Reception**
   - Receive block from peer

2. **Validation**
   - Verify block signature
   - Verify correct producer for block height
   - Verify parent block hash
   - Validate all transactions
   - Verify state root matches

3. **State Update**
   - Execute transaction operations
   - Update in-memory state
   - Persist to BadgerDB

4. **Block Propagation**
   - Forward valid block to other peers

## Node Types

### Producer Nodes

Producer nodes are authorized to create new blocks.

**Characteristics**:
- Require cryptographic keypair
- Must be listed in genesis authorities
- Produce blocks in round-robin order
- Full blockchain validation
- Maintain complete state

**Configuration**:
```yaml
node_type: producer
address: "0xYourAddress"
private_key: "/path/to/key"
```

**Responsibilities**:
- Create blocks when authorized
- Validate all blocks
- Maintain blockchain state
- Serve API requests
- Participate in P2P network

### Full Nodes

Full nodes validate blocks but don't produce them.

**Characteristics**:
- No keypair required
- Validate all blocks
- Maintain complete state
- Cannot create blocks

**Configuration**:
```yaml
node_type: full
```

**Responsibilities**:
- Validate all blocks
- Maintain blockchain state
- Serve API requests
- Participate in P2P network

**Use Cases**:
- Additional API endpoints
- Geographic distribution
- Load balancing
- Backup/redundancy

## Transaction Lifecycle

```
1. Client submits transaction
        ↓
2. API validates format and signature
        ↓
3. Transaction added to mempool
        ↓
4. Producer selects from mempool (when creating block)
        ↓
5. Transaction executed, state updated
        ↓
6. Block created with transaction
        ↓
7. Block broadcast to network
        ↓
8. All nodes validate and apply block
        ↓
9. Transaction confirmed (finalized)
```

## State Management

### State Structure

State is stored as key-value pairs:
- Keys: UTF-8 strings (e.g., "user:alice:name")
- Values: Arbitrary bytes (base64 encoded in API)

### State Operations

**SET**: Create or update a key-value pair
```json
{
  "type": "SET",
  "key": "user:alice:name",
  "value": "QWxpY2U="
}
```

**DELETE**: Remove a key-value pair
```json
{
  "type": "DELETE",
  "key": "user:alice:name"
}
```

### State Root

Each block includes a state root - a Merkle tree hash of all state:
- Enables state verification
- Detects state inconsistencies
- Supports light client proofs (future feature)

## Security Model

### Cryptographic Primitives

- **Hashing**: SHA-256 for block and transaction hashes
- **Signatures**: ECDSA with secp256k1 curve
- **Addresses**: Derived from public keys (Ethereum-compatible)

### Security Guarantees

1. **Block Integrity**: Cryptographic signatures prevent tampering
2. **Transaction Authenticity**: All transactions require valid signatures
3. **Replay Protection**: Nonce prevents transaction replay
4. **Authority Control**: Only authorized producers can create blocks
5. **State Consistency**: Merkle roots ensure state agreement

### Threat Model

**Protected Against**:
- Unauthorized block creation
- Transaction forgery
- Replay attacks
- State manipulation
- Network impersonation

**Not Protected Against** (by design):
- Producer node failures (use multiple producers)
- Network partitions (requires manual intervention)
- Malicious authorities (PoA requires trusted producers)

## Performance Characteristics

### Throughput

- **Block Time**: Configurable (default 5 seconds)
- **Transactions per Block**: Limited by block size (default ~1000 tx)
- **TPS**: Approximately 200 transactions/second

### Latency

- **Transaction Submission**: < 100ms (API response)
- **Block Confirmation**: 1 block time (5 seconds default)
- **Finality**: Immediate (no reorganizations in PoA)

### Storage

- **Block Size**: ~10 KB average
- **Blockchain Growth**: ~172 MB/day (5s blocks, full blocks)
- **State Size**: Depends on application (in-memory + disk)

### Resource Requirements

**Minimum** (Development):
- CPU: 2 cores
- RAM: 2 GB
- Disk: 10 GB
- Network: 1 Mbps

**Recommended** (Production):
- CPU: 4+ cores
- RAM: 8+ GB
- Disk: 100+ GB SSD
- Network: 10+ Mbps

## Scalability

### Horizontal Scaling

- Add more full nodes for API load distribution
- Geographic distribution for lower latency
- Producer nodes don't scale (PoA limitation)

### Vertical Scaling

- Increase block size for more transactions
- Decrease block time for faster confirmation
- More RAM for larger state

### Future Enhancements

- State pruning for reduced disk usage
- Light clients for mobile/embedded devices
- Cross-chain bridges
- Layer 2 solutions

## Code Organization

### Project Structure

```
podoru-chain/
├── cmd/
│   ├── node/              # Node executable
│   └── tools/keygen/      # Key generation tool
├── internal/
│   ├── blockchain/        # Core blockchain logic
│   │   ├── block.go
│   │   ├── chain.go
│   │   └── transaction.go
│   ├── consensus/         # PoA consensus
│   │   └── poa.go
│   ├── crypto/            # Cryptography
│   │   ├── keys.go
│   │   └── signature.go
│   ├── storage/           # BadgerDB integration
│   │   ├── badger.go
│   │   └── state.go
│   ├── network/           # P2P networking
│   │   ├── p2p.go
│   │   ├── peer.go
│   │   └── gossip.go
│   ├── api/rest/          # REST API
│   │   ├── server.go
│   │   └── handlers.go
│   └── node/              # Node orchestration
│       └── node.go
└── config/                # Configuration files
```

### Module Responsibilities

- **blockchain**: Block structure, chain validation, transaction processing
- **consensus**: Authority verification, producer selection
- **crypto**: Key management, signing, verification
- **storage**: Data persistence, state management
- **network**: Peer communication, block propagation
- **api**: HTTP server, request handling
- **node**: Component orchestration, lifecycle management

## Design Principles

1. **Simplicity**: Easy to understand and deploy
2. **Modularity**: Components can be replaced or upgraded
3. **Performance**: Optimized for high throughput
4. **Reliability**: Fault-tolerant and recoverable
5. **Security**: Defense in depth

## Further Reading

- [Consensus (PoA)](consensus.md) - Detailed consensus mechanism
- [Storage (BadgerDB)](storage.md) - Storage layer details
- [P2P Networking](networking.md) - Network architecture
- [Architecture Diagrams](diagrams.md) - Visual architecture reference
