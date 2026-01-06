# Architecture Diagrams

Visual reference for Podoru Chain's architecture and data flows.

## System Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Applications                  │
│                  (Web, Mobile, Desktop, CLI)                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ HTTP/JSON (Port 8545)
                           │
┌──────────────────────────┴──────────────────────────────────┐
│                      REST API Layer                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │   Chain    │  │   State    │  │    Node    │            │
│  │ Endpoints  │  │ Endpoints  │  │ Endpoints  │            │
│  └────────────┘  └────────────┘  └────────────┘            │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────┐
│                   Core Node Components                       │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │              Blockchain Engine                    │      │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │      │
│  │  │  Block   │  │   TX     │  │  State   │       │      │
│  │  │ Manager  │  │ Processor│  │ Manager  │       │      │
│  │  └──────────┘  └──────────┘  └──────────┘       │      │
│  └──────────────────────────────────────────────────┘      │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │           Consensus Engine (PoA)                  │      │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │      │
│  │  │ Producer │  │ Validator│  │  Mempool │       │      │
│  │  │ Selector │  │          │  │          │       │      │
│  │  └──────────┘  └──────────┘  └──────────┘       │      │
│  └──────────────────────────────────────────────────┘      │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │            Network Manager                        │      │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │      │
│  │  │   Peer   │  │  Gossip  │  │   Sync   │       │      │
│  │  │ Manager  │  │ Protocol │  │  Engine  │       │      │
│  │  └──────────┘  └──────────┘  └──────────┘       │      │
│  └──────────────────────────────────────────────────┘      │
│                                                              │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────┐
│                   Storage Layer (BadgerDB)                   │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │   Blocks   │  │   State    │  │    TX      │            │
│  │   Store    │  │   Store    │  │   Store    │            │
│  └────────────┘  └────────────┘  └────────────┘            │
└─────────────────────────────────────────────────────────────┘
                           │
                      Disk Storage
```

## Network Topology

### Multi-Node P2P Network

```
           Producer 1                Producer 2
         (192.168.1.10)            (192.168.1.11)
         API: 8545                 API: 8546
         P2P: 9000                 P2P: 9001
              │   ╲                   ╱   │
              │    ╲                 ╱    │
              │     ╲               ╱     │
              │      ╲             ╱      │
              │       ╲           ╱       │
              │        ╲         ╱        │
              │         ╲       ╱         │
              │          ╲     ╱          │
              │           ╲   ╱           │
              │            ╲ ╱            │
              │             ╳             │
              │            ╱ ╲            │
              │           ╱   ╲           │
              │          ╱     ╲          │
              │         ╱       ╲         │
              │        ╱         ╲        │
              │       ╱           ╲       │
              │      ╱             ╲      │
              │     ╱               ╲     │
              │    ╱                 ╲    │
              │   ╱                   ╲   │
           Producer 3                Full Node 1
         (192.168.1.12)            (192.168.1.13)
         API: 8547                 API: 8548
         P2P: 9002                 P2P: 9003

Legend:
───  TCP P2P Connection (Port 9000-9003)
```

## Block Production Flow

### Producer Node Block Creation

```
┌──────────────────────────────────────────────────────────┐
│ 1. Check if it's producer's turn                         │
│    (height % authority_count == my_index)                │
└────────────────────┬─────────────────────────────────────┘
                     │ Yes
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 2. Wait for block time                                   │
│    (ensure minimum interval since last block)            │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 3. Collect transactions from mempool                     │
│    - Select pending transactions                         │
│    - Validate signatures                                 │
│    - Order by nonce/timestamp                            │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 4. Execute transactions                                  │
│    - Apply SET/DELETE operations                         │
│    - Update in-memory state                              │
│    - Calculate new state root                            │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 5. Create block header                                   │
│    - Height, timestamp, previous hash                    │
│    - Transaction root (Merkle tree)                      │
│    - State root (Merkle tree)                            │
│    - Producer address                                    │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 6. Sign block                                            │
│    - Calculate block hash                                │
│    - Sign with private key (ECDSA)                       │
│    - Add signature to block                              │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 7. Persist block                                         │
│    - Save to BadgerDB                                    │
│    - Update latest height                                │
│    - Persist state changes                               │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 8. Broadcast to network                                  │
│    - Send to all connected peers (gossip)                │
│    - Clear mempool of included transactions              │
└──────────────────────────────────────────────────────────┘
```

## Block Validation Flow

### Full Node Block Reception

```
┌──────────────────────────────────────────────────────────┐
│ 1. Receive block from peer                               │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 2. Check if already seen (duplicate detection)           │
└────────┬────────────────────────────────────────────┬────┘
         │ Already seen                     Not seen  │
         ▼                                            ▼
┌────────────────┐                    ┌────────────────────┐
│ Discard block  │                    │ Continue validation│
└────────────────┘                    └────────┬───────────┘
                                               │
                     ┌─────────────────────────┘
                     ▼
┌──────────────────────────────────────────────────────────┐
│ 3. Verify block signature                                │
│    - Recover address from signature                      │
│    - Compare with expected producer                      │
└────────┬─────────────────────────────────────────────────┘
         │ Valid
         ▼
┌──────────────────────────────────────────────────────────┐
│ 4. Verify block structure                                │
│    - Check height is parent + 1                          │
│    - Verify previous hash matches                        │
│    - Check timestamp > parent timestamp                  │
└────────┬─────────────────────────────────────────────────┘
         │ Valid
         ▼
┌──────────────────────────────────────────────────────────┐
│ 5. Validate all transactions                             │
│    - Verify signatures                                   │
│    - Check nonces                                        │
│    - Validate operation formats                          │
└────────┬─────────────────────────────────────────────────┘
         │ Valid
         ▼
┌──────────────────────────────────────────────────────────┐
│ 6. Execute transactions                                  │
│    - Apply operations to state                           │
│    - Calculate transaction root                          │
│    - Calculate state root                                │
└────────┬─────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────┐
│ 7. Verify Merkle roots                                   │
│    - Compare transaction root                            │
│    - Compare state root                                  │
└────────┬─────────────────────────────────────────────────┘
         │ Valid
         ▼
┌──────────────────────────────────────────────────────────┐
│ 8. Persist block and state                               │
│    - Save block to BadgerDB                              │
│    - Update state store                                  │
│    - Update latest height                                │
└────────┬─────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────┐
│ 9. Propagate to other peers                              │
│    - Forward to connected peers (except sender)          │
│    - Mark block as seen                                  │
└──────────────────────────────────────────────────────────┘
```

## Transaction Lifecycle

### From Submission to Confirmation

```
Client Application
       │
       │ 1. Create transaction
       │    (operations, nonce)
       │
       ▼
   Sign with private key
       │
       │ 2. Submit via API
       │    POST /api/v1/transaction
       │
       ▼
┌────────────────────┐
│    API Server      │
│  - Validate format │
│  - Verify signature│
└─────────┬──────────┘
          │ 3. Add to mempool
          ▼
┌────────────────────┐
│     Mempool        │
│  - Pending TXs     │
│  - Ordered by time │
└─────────┬──────────┘
          │
          │ 4. Producer selects TX
          │    (when creating block)
          ▼
┌────────────────────┐
│  Block Creation    │
│  - Execute TX      │
│  - Update state    │
└─────────┬──────────┘
          │
          │ 5. Block propagated
          ▼
┌────────────────────┐
│  All Nodes         │
│  - Validate block  │
│  - Apply TX        │
└─────────┬──────────┘
          │
          │ 6. Block finalized
          ▼
    TX Confirmed
  (Permanent in blockchain)
```

## State Management

### State Storage Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Application Layer                      │
│               (REST API, Transaction Processor)           │
└────────────────────────┬─────────────────────────────────┘
                         │
                         │ Read/Write Operations
                         │
┌────────────────────────┴─────────────────────────────────┐
│                    State Manager                          │
│                                                           │
│  ┌────────────────────────────────────────────────┐     │
│  │           In-Memory State Cache                 │     │
│  │                                                 │     │
│  │   map[string][]byte {                          │     │
│  │     "user:alice:name" → "Alice"                │     │
│  │     "user:alice:email" → "alice@example.com"   │     │
│  │     "post:123:content" → "Hello World"         │     │
│  │     ...                                        │     │
│  │   }                                            │     │
│  └────────────────┬───────────────────────────────┘     │
│                   │                                      │
│                   │ Cache miss → Read from disk         │
│                   │ Write → Update cache + disk         │
│                   │                                      │
└───────────────────┴──────────────────────────────────────┘
                    │
                    │ Persist
                    │
┌───────────────────┴──────────────────────────────────────┐
│                   BadgerDB Storage                        │
│                                                           │
│  ┌─────────────────────────────────────────────┐         │
│  │  Key-Value Store                            │         │
│  │                                             │         │
│  │  state:user:alice:name → "Alice"           │         │
│  │  state:user:alice:email → "alice@..."      │         │
│  │  state:post:123:content → "Hello..."       │         │
│  │  ...                                       │         │
│  └─────────────────────────────────────────────┘         │
└──────────────────────────────────────────────────────────┘
                    │
                    ▼
              Disk Storage
```

## Consensus Timeline

### Round-Robin Block Production

```
Time ─────────────────────────────────────────────────────▶

Block  0        5s       10s       15s       20s       25s
Height │        │        │         │         │         │
       │        │        │         │         │         │
   0   █
       Producer 1
       │
   1            █
                Producer 2
                │
   2                     █
                         Producer 3
                         │
   3                              █
                                  Producer 1 (wraps)
                                  │
   4                                       █
                                           Producer 2
                                           │
   5                                                █
                                                   Producer 3

Legend:
█  Block created
│  Block propagation time (~500ms)
```

## Data Flow Diagram

### Complete Data Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ HTTP Request
       │ POST /api/v1/transaction
       │
       ▼
┌─────────────────────────────────────────┐
│           REST API Handler              │
│                                         │
│  1. Validate JSON format                │
│  2. Verify transaction signature        │
│  3. Check nonce                         │
└──────┬──────────────────────────────────┘
       │
       │ Valid transaction
       │
       ▼
┌─────────────────────────────────────────┐
│            Mempool                      │
│                                         │
│  Pending transactions queue             │
└──────┬──────────────────────────────────┘
       │
       │ Producer selects TXs
       │ (when it's their turn)
       │
       ▼
┌─────────────────────────────────────────┐
│       Transaction Processor             │
│                                         │
│  1. Execute SET/DELETE operations       │
│  2. Update state manager                │
│  3. Calculate state root                │
└──────┬──────────────────────────────────┘
       │
       │ State changes
       │
       ▼
┌─────────────────────────────────────────┐
│         State Manager                   │
│                                         │
│  1. Update in-memory cache              │
│  2. Persist to BadgerDB                 │
└──────┬──────────────────────────────────┘
       │
       │ State persisted
       │
       ▼
┌─────────────────────────────────────────┐
│        Block Creator                    │
│                                         │
│  1. Build block with transactions       │
│  2. Add Merkle roots                    │
│  3. Sign block                          │
└──────┬──────────────────────────────────┘
       │
       │ New block
       │
       ▼
┌─────────────────────────────────────────┐
│        Block Store                      │
│                                         │
│  1. Save to BadgerDB                    │
│  2. Index by height and hash            │
└──────┬──────────────────────────────────┘
       │
       │ Block stored
       │
       ▼
┌─────────────────────────────────────────┐
│       Network Manager                   │
│                                         │
│  Broadcast block to all peers           │
└─────────────────────────────────────────┘
```

## Legend

### Symbol Reference

```
┌─────┐
│ Box │  Component or process
└─────┘

   │
   ▼     Flow direction

   ╱╲
  ╱  ╲   Connection/relationship
 ╱    ╲

   ─     Direct connection

   █     Active process/block
```

## Further Reading

- [Architecture Overview](README.md)
- [Consensus Details](consensus.md)
- [Storage Details](storage.md)
- [Networking Details](networking.md)
