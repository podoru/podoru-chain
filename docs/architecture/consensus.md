# Consensus (Proof of Authority)

Podoru Chain uses Proof of Authority (PoA) consensus, a deterministic and efficient consensus mechanism ideal for permissioned and semi-permissioned blockchains.

## Overview

Proof of Authority relies on a fixed set of authorized nodes (authorities) that take turns producing blocks in a round-robin fashion.

### Key Characteristics

- **Deterministic**: No randomness in block producer selection
- **Efficient**: Low computational overhead (no mining)
- **Fast**: Immediate finality, no forks
- **Permissioned**: Only authorized addresses can produce blocks

## How PoA Works

### 1. Authority Set

Authorities are defined in the genesis block and remain fixed:

```json
{
  "authorities": [
    "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
    "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
  ]
}
```

**Properties**:
- Fixed set (no dynamic changes in current version)
- Minimum 1 authority (not recommended for production)
- Recommended 3+ authorities for fault tolerance
- Maximum 10 authorities (practical limit)

### 2. Round-Robin Selection

Block producers are selected using a simple formula:

```
producer_index = block_height % authority_count
producer = authorities[producer_index]
```

**Example** with 3 authorities:
- Block 0: Authority 0
- Block 1: Authority 1
- Block 2: Authority 2
- Block 3: Authority 0 (wraps around)
- Block 4: Authority 1
- ...

### 3. Block Production

When it's a producer's turn:

1. **Check Turn**
   ```go
   expectedProducer := authorities[height % len(authorities)]
   if myAddress != expectedProducer {
       return // Not my turn
   }
   ```

2. **Wait for Block Time**
   ```go
   timeSinceLastBlock := now - lastBlock.Timestamp
   if timeSinceLastBlock < blockTime {
       sleep(blockTime - timeSinceLastBlock)
   }
   ```

3. **Create Block**
   - Collect transactions from mempool
   - Execute transactions and update state
   - Calculate Merkle roots
   - Build block header

4. **Sign Block**
   ```go
   hash := calculateBlockHash(block)
   signature := sign(hash, privateKey)
   block.Signature = signature
   ```

5. **Broadcast**
   - Send to all connected peers
   - Clear mempool of included transactions

### 4. Block Validation

When any node receives a block:

1. **Verify Producer**
   ```go
   expectedProducer := authorities[block.Height % len(authorities)]
   recoveredAddress := recoverAddress(block.Hash, block.Signature)

   if recoveredAddress != expectedProducer {
       return ErrInvalidProducer
   }
   ```

2. **Verify Signature**
   ```go
   if !verifySignature(block.Hash, block.Signature, block.Producer) {
       return ErrInvalidSignature
   }
   ```

3. **Verify Timestamp**
   ```go
   if block.Timestamp <= parentBlock.Timestamp {
       return ErrInvalidTimestamp
   }
   ```

4. **Validate Transactions**
   - Check all transaction signatures
   - Verify nonces
   - Execute operations

5. **Verify State Root**
   ```go
   calculatedStateRoot := calculateMerkleRoot(newState)
   if calculatedStateRoot != block.StateRoot {
       return ErrInvalidStateRoot
   }
   ```

## Block Structure

### Block Header

```go
type Block struct {
    Height          uint64     // Block number
    Timestamp       int64      // Unix timestamp
    PreviousHash    string     // Hash of parent block
    Producer        string     // Address of block producer
    TransactionRoot string     // Merkle root of transactions
    StateRoot       string     // Merkle root of state
    Transactions    []Transaction
    Signature       string     // Producer's signature
}
```

### Block Hash Calculation

```go
func calculateBlockHash(block *Block) string {
    data := fmt.Sprintf(
        "%d:%d:%s:%s:%s:%s",
        block.Height,
        block.Timestamp,
        block.PreviousHash,
        block.Producer,
        block.TransactionRoot,
        block.StateRoot,
    )
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

## Timing and Synchronization

### Block Time

Configurable interval between blocks (default: 5 seconds):

```yaml
block_time: 5s
```

**Considerations**:
- Shorter = faster confirmation, more network overhead
- Longer = slower confirmation, more efficient
- Recommended: 3-10 seconds for most applications

### Clock Synchronization

Nodes should have synchronized clocks (NTP recommended):
- Prevents timestamp conflicts
- Ensures predictable block production
- Allows proper block time enforcement

### Late Blocks

If a producer doesn't produce a block within block_time:

1. Next producer waits for `2 * block_time`
2. If still no block, produces block anyway
3. Network continues with next authority

## Fault Tolerance

### Producer Failures

With 3+ authorities, the network tolerates failures:

**Example with 3 authorities**:
- 1 producer fails: Network continues (66% capacity)
- 2 producers fail: Network halts (< 50% capacity)

**Calculation**:
```
Fault tolerance = floor((N - 1) / 2)

N=3: Tolerates 1 failure
N=5: Tolerates 2 failures
N=7: Tolerates 3 failures
```

### Recovery

Failed producer node can rejoin:
1. Sync blockchain from peers
2. Resume producing when it's their turn
3. No manual intervention needed

## Security

### Attack Vectors

**Unauthorized Block Production**:
- Prevention: Signature verification
- Each block must be signed by expected authority
- Invalid signatures rejected immediately

**Block Withholding**:
- Impact: Temporary slowdown (other producers continue)
- Mitigation: Use 3+ producers for redundancy

**Producer Collusion**:
- Impact: Authorities could censor transactions
- Mitigation: Choose trustworthy authorities
- Future: Dynamic authority set with voting

**Network Partitions**:
- Impact: Could create temporary splits
- Resolution: Manual intervention, longest chain wins
- PoA provides deterministic finality (no true forks)

### Signature Scheme

**Algorithm**: ECDSA with secp256k1 curve
- Same as Ethereum for compatibility
- Well-tested and widely used
- Recoverable signatures (can derive public key)

**Address Derivation**:
```
PublicKey → Keccak256 → Last 20 bytes → Address
```

## Finality

### Immediate Finality

PoA provides immediate finality:
- No block reorganizations
- Once validated, blocks never change
- Transactions confirmed in 1 block

**Contrast with PoW**:
- PoW: Probabilistic finality (wait for confirmations)
- PoA: Deterministic finality (instant)

### Fork Resolution

True forks shouldn't occur in PoA, but if they do:
1. Nodes reject blocks from wrong producer
2. Only one valid chain exists
3. Invalid chains automatically rejected

## Performance

### Throughput

With default settings:
- Block time: 5 seconds
- Transactions per block: ~1000
- TPS: ~200 transactions/second

### Latency

- Transaction submission: < 100ms
- Block creation: 5 seconds (configurable)
- Finality: Immediate (1 block)

### Scalability

**Vertical**:
- Decrease block time for faster confirmation
- Increase block size for more transactions

**Horizontal**:
- Adding authorities doesn't increase throughput
- Diminishing returns after 5-7 authorities
- More authorities = more network overhead

## Configuration

### Consensus Parameters

```yaml
# Consensus configuration
authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"

block_time: 5s
```

### Best Practices

**Number of Authorities**:
- Minimum: 1 (development only)
- Recommended: 3-5 (production)
- Maximum: 10 (practical limit)

**Block Time**:
- Development: 3-5 seconds
- Production: 5-10 seconds
- High-throughput: 2-3 seconds

**Authority Selection**:
- Choose trustworthy entities
- Geographic distribution
- Different organizations
- Redundant infrastructure

## Comparison with Other Consensus

### vs Proof of Work (PoW)

| Feature | PoA | PoW |
|---------|-----|-----|
| Energy | Minimal | High |
| Speed | Fast (5s) | Slow (10+ min) |
| Finality | Immediate | Probabilistic |
| Throughput | High | Low |
| Decentralization | Lower | Higher |
| Cost | Low | High |

### vs Proof of Stake (PoS)

| Feature | PoA | PoS |
|---------|-----|-----|
| Capital requirement | None | Stake required |
| Validator selection | Fixed | Stake-weighted |
| Complexity | Simple | Complex |
| Security model | Trust | Economic |
| Sybil resistance | Authority list | Stake |

### vs Practical Byzantine Fault Tolerance (PBFT)

| Feature | PoA (Podoru) | PBFT |
|---------|--------------|------|
| Communication | Gossip | All-to-all |
| Complexity | Simple | Complex |
| Latency | Low | Medium |
| Fault tolerance | (N-1)/2 | (N-1)/3 |
| Scalability | Good | Limited |

## Use Cases

PoA is ideal for:

- **Private/Consortium Chains**: Known set of validators
- **Application-Specific Chains**: Single organization control
- **Development Networks**: Fast iteration, low cost
- **Semi-Public Chains**: Trusted authorities, open participation

PoA is NOT ideal for:

- **Fully Decentralized Networks**: No trusted authorities
- **High Adversarial Environments**: Authority collusion risk
- **Permissionless Systems**: Open validator set required

## Future Enhancements

### Dynamic Authority Set

Allow authorities to be added/removed via governance:
```go
// Future feature
AddAuthority(address, votes)
RemoveAuthority(address, votes)
```

### Slashing

Penalize misbehaving authorities:
- Missed blocks
- Invalid blocks
- Downtime

### Checkpointing

Periodic checkpoints for long-term finality:
- Super-majority signatures
- External verification
- Cross-chain proofs

## Monitoring

### Metrics to Track

**Producer Health**:
```bash
# Block production rate
curl http://localhost:8545/api/v1/chain/info | jq '.data.height'

# Missed blocks (compare expected vs actual height)
expected_height = (current_time - genesis_time) / block_time
```

**Network Consensus**:
```bash
# Check all nodes have same height
for port in 8545 8546 8547 8548; do
  curl -s http://localhost:$port/api/v1/chain/info | jq '.data.height'
done
```

**Timing**:
```bash
# Block intervals
curl http://localhost:8545/api/v1/block/latest | jq '.data.timestamp'
```

## Troubleshooting

### Blocks Not Being Produced

1. Check producer node is running
2. Verify correct address in configuration
3. Check clock synchronization
4. Review logs for errors

### Fork Detected

1. Check all nodes have same genesis file
2. Verify authority list matches
3. Check for network partitions
4. Restart affected nodes

### Slow Block Production

1. Check network connectivity
2. Verify clock synchronization
3. Monitor producer node resources
4. Check block time configuration

## Further Reading

- [Architecture Overview](README.md)
- [Storage Layer](storage.md)
- [P2P Networking](networking.md)
- [Configuration Guide](../configuration/README.md)
