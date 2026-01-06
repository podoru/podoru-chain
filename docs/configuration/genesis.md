# Genesis Block Configuration

The genesis block is the first block in the blockchain and defines the initial state of the network.

## Overview

The genesis configuration file:
- Defines initial blockchain authorities
- Sets initial state (key-value pairs)
- Establishes network parameters
- Must be identical across all nodes

**Format**: JSON
**Location**: Typically `genesis.json` or specified in node config

## File Structure

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
    "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
  ],
  "initial_state": {
    "chain:name": "Podoru Chain",
    "chain:version": "1.0.0",
    "chain:description": "Decentralized blockchain for storing any data",
    "system:initialized": "true"
  }
}
```

## Parameters

### timestamp

**Type**: Integer (Unix timestamp)
**Required**: Yes
**Description**: Creation time of genesis block

```json
"timestamp": 1704556800
```

**Recommendations**:
- Use current time: `date +%s`
- Or use fixed historical time
- All nodes should use same timestamp

### authorities

**Type**: Array of strings
**Required**: Yes
**Description**: List of authorized block producer addresses

```json
"authorities": [
  "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
  "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
  "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
]
```

**Requirements**:
- Must be Ethereum-compatible addresses (0x-prefixed, 42 characters)
- At least 1 authority required
- Recommended: 3+ authorities for production
- Maximum: 10 authorities (practical limit)
- Must match producer node addresses in config files

**Obtaining Addresses**:
```bash
./bin/keygen -output producer1.key
# Output includes: Address: 0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB
```

### initial_state

**Type**: Object (key-value pairs)
**Required**: No (can be empty object)
**Description**: Initial blockchain state

```json
"initial_state": {
  "chain:name": "My Custom Chain",
  "chain:version": "1.0.0",
  "chain:description": "A custom blockchain",
  "system:initialized": "true",
  "custom:key": "custom value"
}
```

**Common Keys**:
- `chain:name` - Blockchain name
- `chain:version` - Version identifier
- `chain:description` - Description
- `system:initialized` - Initialization flag

**Custom State**:
You can add any initial key-value pairs:

```json
"initial_state": {
  "chain:name": "My Chain",
  "admin:address": "0xAdminAddress",
  "config:max_tx_size": "1048576",
  "feature:enabled": "true"
}
```

## Examples

### Minimal Genesis

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  ],
  "initial_state": {}
}
```

### Development Genesis

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0xProducer1Address",
    "0xProducer2Address",
    "0xProducer3Address"
  ],
  "initial_state": {
    "chain:name": "Podoru Dev",
    "chain:version": "1.0.0-dev",
    "chain:description": "Development blockchain",
    "env:mode": "development"
  }
}
```

### Production Genesis

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0xProducer1Address",
    "0xProducer2Address",
    "0xProducer3Address",
    "0xProducer4Address",
    "0xProducer5Address"
  ],
  "initial_state": {
    "chain:name": "Podoru Mainnet",
    "chain:version": "1.0.0",
    "chain:description": "Production blockchain for decentralized data storage",
    "system:initialized": "true",
    "network:id": "podoru-mainnet-1"
  }
}
```

## Creating Genesis File

### Using Setup Wizard

The setup wizard automatically creates genesis file:

```bash
make setup-wizard
```

### Manual Creation

1. **Generate producer keys**:
```bash
./bin/keygen -output keys/producer1.key  # Note the address
./bin/keygen -output keys/producer2.key
./bin/keygen -output keys/producer3.key
```

2. **Create genesis.json**:
```bash
cat > genesis.json <<'EOF'
{
  "timestamp": $(date +%s),
  "authorities": [
    "0xAddress1FromKeygen",
    "0xAddress2FromKeygen",
    "0xAddress3FromKeygen"
  ],
  "initial_state": {
    "chain:name": "My Chain",
    "chain:version": "1.0.0"
  }
}
EOF
```

3. **Distribute to all nodes**:
```bash
# All nodes must have IDENTICAL genesis file
cp genesis.json docker/data/producer1/
cp genesis.json docker/data/producer2/
cp genesis.json docker/data/producer3/
cp genesis.json docker/data/fullnode1/
```

## Validation

### Check JSON Syntax

```bash
cat genesis.json | jq .
```

### Verify Authorities

```bash
# Extract authorities
cat genesis.json | jq '.authorities[]'

# Count authorities
cat genesis.json | jq '.authorities | length'
```

### Verify Address Format

```bash
# Check all addresses are valid
cat genesis.json | jq -r '.authorities[]' | while read addr; do
  if [[ ! $addr =~ ^0x[0-9a-fA-F]{40}$ ]]; then
    echo "Invalid address: $addr"
  fi
done
```

## Consistency Requirements

**Critical**: All nodes must use the EXACT SAME genesis file.

### Same Timestamp

```bash
# ✅ Good: All nodes use same genesis
producer1/genesis.json: "timestamp": 1704556800
producer2/genesis.json: "timestamp": 1704556800
producer3/genesis.json: "timestamp": 1704556800

# ❌ Bad: Different timestamps
producer1/genesis.json: "timestamp": 1704556800
producer2/genesis.json: "timestamp": 1704556900  # Different!
```

### Same Authorities

```bash
# ✅ Good: Same order, same addresses
producer1/genesis.json: ["0xA", "0xB", "0xC"]
producer2/genesis.json: ["0xA", "0xB", "0xC"]

# ❌ Bad: Different order
producer1/genesis.json: ["0xA", "0xB", "0xC"]
producer2/genesis.json: ["0xB", "0xA", "0xC"]  # Different order!
```

### Verification

```bash
# Compute hash of genesis file
sha256sum genesis.json

# All nodes should have same hash
# producer1: a1b2c3...
# producer2: a1b2c3...  # Same
# producer3: a1b2c3...  # Same
```

## Modifying Genesis

### Before Network Start

You can freely modify genesis before starting nodes.

```bash
# Edit genesis
vi genesis.json

# Copy to all nodes
for node in producer1 producer2 producer3 fullnode1; do
  cp genesis.json docker/data/$node/
done
```

### After Network Start

**Cannot modify** genesis after network is running.

To change:
1. Stop all nodes
2. Delete blockchain data
3. Update genesis file
4. Restart network (new blockchain)

**Warning**: This destroys all existing blockchain data!

## Troubleshooting

### Genesis Hash Mismatch

If nodes can't sync:

```bash
# Check genesis hash on each node
curl http://localhost:8545/api/v1/chain/info | jq '.data.genesis_hash'
curl http://localhost:8546/api/v1/chain/info | jq '.data.genesis_hash'

# If different, nodes have different genesis files
```

**Solution**: Ensure all nodes have identical genesis.json

### Invalid Authority Address

```
Error: invalid authority address: 0xInvalidAddress
```

**Solution**: Verify all authorities are valid Ethereum addresses

### JSON Syntax Error

```
Error: invalid character '}' looking for beginning of value
```

**Solution**: Validate JSON syntax:
```bash
cat genesis.json | jq .
```

## Further Reading

- [Producer Configuration](producer.md) - Configure producer nodes
- [Full Node Configuration](fullnode.md) - Configure full nodes
- [Manual Setup Guide](../getting-started/manual-setup.md) - Complete setup
