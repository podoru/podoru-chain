# P2P Networking

Podoru Chain uses a TCP-based peer-to-peer network with a gossip protocol for efficient block and transaction propagation.

## Overview

The P2P networking layer provides:
- Peer discovery and connection management
- Block and transaction propagation
- Blockchain synchronization
- Network health monitoring

## Network Architecture

```
Node A (Producer)
    ↓ ↑
    TCP Connections
    ↓ ↑
┌───┴───┬───────┬───────┐
│       │       │       │
Node B  Node C  Node D  Node E
(Prod)  (Prod)  (Full)  (Full)
│       │       │       │
└───┬───┴───┬───┴───┬───┘
    │       │       │
    └───────┴───────┘
    (Fully Connected Mesh)
```

## Protocol Design

### Message Types

```go
type MessageType uint8

const (
    MessageTypeBlock         MessageType = 1
    MessageTypeTransaction   MessageType = 2
    MessageTypeBlockRequest  MessageType = 3
    MessageTypeBlockResponse MessageType = 4
    MessageTypePeerList      MessageType = 5
    MessageTypeHandshake     MessageType = 6
)
```

### Message Format

```go
type Message struct {
    Type      MessageType    // Message type
    Payload   []byte         // Serialized data
    Timestamp int64          // Unix timestamp
    Signature string         // Optional signature
}
```

**Serialization**: Messages are JSON-encoded for simplicity.

```json
{
  "type": 1,
  "payload": "base64-encoded-data",
  "timestamp": 1704556800,
  "signature": "0x..."
}
```

## Connection Management

### Bootstrap Process

1. **Load Bootstrap Peers**
   ```yaml
   bootstrap_peers:
     - "192.168.1.10:9000"
     - "192.168.1.11:9001"
     - "192.168.1.12:9002"
   ```

2. **Connect to Peers**
   ```go
   for _, peerAddr := range config.BootstrapPeers {
       go connectToPeer(peerAddr)
   }
   ```

3. **Handshake**
   ```go
   type Handshake struct {
       Version     string
       ChainID     string
       Height      uint64
       NodeType    string
       ListenAddr  string
   }
   ```

4. **Discover More Peers**
   - Request peer list from connected peers
   - Connect to new peers
   - Maintain max peer limit

### Peer Discovery

**Gossip-Based Discovery**:
```go
func requestPeerList(peer *Peer) {
    msg := &Message{
        Type: MessageTypePeerList,
    }
    peer.Send(msg)
}

func handlePeerList(peers []string) {
    for _, addr := range peers {
        if !isConnected(addr) && peerCount < maxPeers {
            connectToPeer(addr)
        }
    }
}
```

**No External Discovery**:
- No DHT or centralized registry
- Relies on bootstrap peers
- Simple and predictable

### Connection Lifecycle

```
1. TCP Connect
     ↓
2. Send Handshake
     ↓
3. Receive Handshake
     ↓
4. Validate (version, chain ID)
     ↓
5. Add to Peer List
     ↓
6. Start Message Loops
     ↓
7. Exchange Data
     ↓
8. Disconnect (on error/shutdown)
```

### Peer Management

```go
type PeerManager struct {
    peers      map[string]*Peer
    maxPeers   int
    mutex      sync.RWMutex
}

type Peer struct {
    Address    string
    Conn       net.Conn
    Version    string
    Height     uint64
    LastSeen   time.Time
}
```

**Maximum Peers**:
```yaml
max_peers: 50  # Default
```

**Peer Selection**:
- Prioritize bootstrap peers
- Maintain diverse peer set
- Replace inactive peers
- Limit connections per IP (future)

## Block Propagation

### Gossip Protocol

When a producer creates a new block:

1. **Broadcast to All Peers**
   ```go
   func broadcastBlock(block *Block) {
       blockData, _ := json.Marshal(block)
       msg := &Message{
           Type:    MessageTypeBlock,
           Payload: blockData,
       }

       for _, peer := range peers {
           peer.Send(msg)
       }
   }
   ```

2. **Receive and Validate**
   ```go
   func handleBlock(block *Block, fromPeer *Peer) {
       // Validate block
       if err := validateBlock(block); err != nil {
           return
       }

       // Add to blockchain
       blockchain.AddBlock(block)

       // Propagate to other peers (except sender)
       relayBlock(block, fromPeer)
   }
   ```

3. **Relay to Other Peers**
   ```go
   func relayBlock(block *Block, excludePeer *Peer) {
       for _, peer := range peers {
           if peer != excludePeer {
               peer.Send(blockMessage)
           }
       }
   }
   ```

### Duplicate Detection

```go
type BlockCache struct {
    seen  map[string]bool
    mutex sync.RWMutex
}

func (bc *BlockCache) HasSeen(hash string) bool {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    return bc.seen[hash]
}

func (bc *BlockCache) MarkSeen(hash string) {
    bc.mutex.Lock()
    defer bc.mutex.Unlock()
    bc.seen[hash] = true
}
```

**Benefits**:
- Prevents redundant processing
- Reduces network traffic
- Breaks propagation loops

## Transaction Propagation

Similar to blocks but with mempool:

1. **Receive Transaction**
   ```go
   func handleTransaction(tx *Transaction) {
       // Validate
       if err := validateTransaction(tx); err != nil {
           return
       }

       // Add to mempool
       mempool.Add(tx)

       // Broadcast to peers
       broadcastTransaction(tx)
   }
   ```

2. **Broadcast**
   ```go
   func broadcastTransaction(tx *Transaction) {
       txData, _ := json.Marshal(tx)
       msg := &Message{
           Type:    MessageTypeTransaction,
           Payload: txData,
       }

       for _, peer := range peers {
           peer.Send(msg)
       }
   }
   ```

## Blockchain Synchronization

### Sync Process

When a node starts or falls behind:

1. **Detect Height Gap**
   ```go
   func checkSync() {
       for _, peer := range peers {
           if peer.Height > myHeight + 10 {
               startSync(peer)
               break
           }
       }
   }
   ```

2. **Request Missing Blocks**
   ```go
   func syncBlocks(fromHeight, toHeight uint64, peer *Peer) {
       for h := fromHeight; h <= toHeight; h++ {
           requestBlock(h, peer)
       }
   }

   func requestBlock(height uint64, peer *Peer) {
       msg := &Message{
           Type:    MessageTypeBlockRequest,
           Payload: []byte(fmt.Sprintf("%d", height)),
       }
       peer.Send(msg)
   }
   ```

3. **Receive and Validate Blocks**
   ```go
   func handleBlockResponse(block *Block) {
       if err := validateBlock(block); err != nil {
           log.Printf("Invalid block during sync: %v", err)
           return
       }

       blockchain.AddBlock(block)

       if blockchain.Height < targetHeight {
           requestBlock(blockchain.Height + 1, syncPeer)
       }
   }
   ```

4. **Apply Blocks Sequentially**
   - Validate each block
   - Execute transactions
   - Update state
   - Continue until caught up

### Batch Synchronization

For faster sync, request blocks in batches:

```go
func syncBatch(startHeight, endHeight uint64, peer *Peer) {
    batchSize := 100
    for h := startHeight; h <= endHeight; h += batchSize {
        end := min(h + batchSize - 1, endHeight)
        requestBlockRange(h, end, peer)
    }
}
```

## Network Configuration

### Configuration Options

```yaml
# P2P Network Settings
p2p_port: 9000                    # Port to listen on
p2p_bind_addr: "0.0.0.0"          # Bind address
bootstrap_peers:                   # Initial peers
  - "192.168.1.10:9000"
  - "192.168.1.11:9001"
max_peers: 50                      # Maximum connections
```

### Port Configuration

**Default Ports**:
- P2P: 9000
- API: 8545

**Firewall Rules**:
```bash
# Allow P2P connections
sudo ufw allow 9000/tcp

# Allow API (optional, if public)
sudo ufw allow 8545/tcp
```

## Security

### Connection Security

**Authentication**:
- Handshake validates chain ID and version
- Future: TLS encryption for connections
- Future: Node authentication via signatures

**DoS Protection**:
```go
// Rate limiting
type RateLimiter struct {
    limits map[string]*rate.Limiter
}

func (rl *RateLimiter) Allow(peerAddr string) bool {
    limiter := rl.getLimiter(peerAddr)
    return limiter.Allow()
}
```

**Limits**:
- Max connections per IP: 3 (future)
- Max message size: 10 MB
- Max messages per second: 100

### Validation

All received data is validated:

**Block Validation**:
- Signature verification
- Authority check
- State root verification
- Transaction validation

**Transaction Validation**:
- Signature verification
- Nonce check
- Format validation

## Performance

### Latency

**Block Propagation**:
- 1 hop: < 100ms
- 3 hops: < 300ms
- Full network: < 1 second

**Transaction Propagation**:
- Similar to block propagation
- Mempool sync: continuous

### Throughput

**Network Bandwidth**:
- Block size: ~10 KB average
- Block frequency: 5 seconds
- Bandwidth: ~2 KB/s per peer

**Scaling**:
- More peers = faster propagation
- More peers = more bandwidth
- Optimal: 5-10 well-connected peers

## Monitoring

### Network Health

```bash
# Check connected peers
curl http://localhost:8545/api/v1/node/peers

# Check node info
curl http://localhost:8545/api/v1/node/info
```

**Response**:
```json
{
  "success": true,
  "data": {
    "peers": [
      {
        "address": "192.168.1.10:9000",
        "height": 1234,
        "last_seen": 1704556800
      }
    ],
    "peer_count": 3
  }
}
```

### Metrics

**Key Metrics**:
- Peer count
- Block propagation time
- Sync status
- Network bandwidth

**Logging**:
```go
log.Printf("Block %d propagated to %d peers in %dms",
    block.Height, peerCount, latency)
```

## Troubleshooting

### No Peers Connecting

1. **Check bootstrap peers**
   ```bash
   # Test connectivity
   nc -zv 192.168.1.10 9000
   ```

2. **Verify firewall**
   ```bash
   sudo ufw status
   ```

3. **Check logs**
   ```bash
   docker logs podoru-producer1 | grep "peer"
   ```

### Slow Synchronization

1. **Check peer heights**
   ```bash
   curl http://localhost:8545/api/v1/node/peers
   ```

2. **Verify network bandwidth**
   ```bash
   iftop -i eth0
   ```

3. **Check for errors**
   ```bash
   docker logs podoru-producer1 | grep "sync"
   ```

### Network Partitions

Symptoms:
- Nodes have different heights
- Blocks not propagating

Resolution:
1. Check connectivity between all nodes
2. Verify same genesis file
3. Restart affected nodes
4. Check for network issues

## Best Practices

### Network Topology

**For Production**:
- Use 3-5 bootstrap peers
- Geographic distribution
- Redundant connections
- Monitor peer health

**For Development**:
- Use localhost for all nodes
- Minimal peers (1-2)
- No external connectivity needed

### Peer Selection

**Good Peers**:
- Low latency
- High uptime
- Correct chain
- Recent blocks

**Bad Peers**:
- High latency
- Frequent disconnects
- Wrong chain
- Stale blocks

## Future Enhancements

### TLS Encryption

Encrypt P2P connections:
```go
// Future feature
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
}
conn, _ := tls.Dial("tcp", addr, tlsConfig)
```

### Node Discovery

DHT-based discovery:
```go
// Future feature
dht := NewDHT()
peers := dht.FindPeers()
```

### Advanced Routing

- Kademlia-based routing
- Geographic awareness
- Load balancing

## Further Reading

- [Architecture Overview](README.md)
- [Consensus Mechanism](consensus.md)
- [Configuration Guide](../configuration/README.md)
- [Troubleshooting](../troubleshooting/README.md)
